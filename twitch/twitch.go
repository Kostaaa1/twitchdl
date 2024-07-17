package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	file "github.com/Kostaaa1/twitchdl/utils"
)

const (
	gqlURL      = "https://gql.twitch.tv/gql"
	gqlClientID = "kimne78kx3ncx6brgo4mv6wki5h1ko"
	usherURL    = "https://usher.ttvnw.net"
	helixURL    = "https://api.twitch.tv/helix"
)

type Client struct {
	client      *http.Client
	gqlURL      string
	helixURL    string
	usherURL    string
	gqlClientID string
	mu          sync.Mutex
}

type VideoType int

const (
	TypeClip VideoType = iota
	TypeVOD
	TypeLivestream
)

// change the name
func (c *Client) Name(vType VideoType, id string) (string, error) {
	var name string
	switch vType {
	case TypeClip:
		clip, err := c.ClipMetadata(id)
		if err != nil {
			return "", err
		}
		name = fmt.Sprintf("%s - %s", clip.Broadcaster.DisplayName, clip.Title)
	case TypeVOD:
		vod, err := c.VideoMetadata(id)
		if err != nil {
			fmt.Println("error", err)
			return "", err
		}
		name = fmt.Sprintf("%s - %s", vod.Owner.Login, vod.Title)
	}
	return name, nil
}

// change the name
func (c *Client) PathName(vType VideoType, id, output string) (string, error) {
	name, err := c.Name(vType, id)
	if err != nil {
		return "", err
	}
	name = file.CreateVideo(output, name)
	return name, nil
}

func (c *Client) ID(URL string) (string, VideoType, error) {
	u, err := url.Parse(URL)
	s := strings.Split(u.Path, "/")
	if len(s) == 2 {
		return s[1], TypeLivestream, nil
	}
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse the URL: %s", err)
	}
	if !strings.Contains(u.Hostname(), "twitch.tv") {
		return "", 0, fmt.Errorf("the hostname of the URL does not contain twitch.tv")
	}
	if strings.Contains(u.Path, "/clip/") {
		_, id := path.Split(u.Path)
		return id, TypeClip, nil
	}
	if strings.Contains(u.Path, "/videos/") {
		_, id := path.Split(u.Path)
		return id, TypeVOD, nil
	}
	return "", 0, fmt.Errorf("failed to get the information from the URL")
}

func New(client *http.Client) Client {
	return Client{
		client:      client,
		gqlURL:      gqlURL,
		gqlClientID: gqlClientID,
		usherURL:    usherURL,
		helixURL:    helixURL,
		mu:          sync.Mutex{},
	}
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %s", err)
	}
	if s := resp.StatusCode; s < 200 || s >= 300 {
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status code %d: %s", s, string(b))
	}
	return resp, nil
}

func (c *Client) Fetch(url string) ([]byte, error) {
	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching m3u8 failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("non-success HTTP status: %d %s", resp.StatusCode, resp.Status)
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %w", err)
	}
	return bytes, nil
}

func (c *Client) decodeJSONResponse(resp *http.Response, p interface{}) error {
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return err
	}
	return nil
}

func (c *Client) readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *Client) sendGraphqlLoadAndDecode(body *strings.Reader, v any) error {
	req, err := http.NewRequest(http.MethodPost, c.gqlURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request to get the access token: %s", err)
	}
	req.Header.Set("Client-Id", c.gqlClientID)

	resp, err := c.do(req)
	if err != nil {
		fmt.Println("DO ERRPR ", err)
		return err
	}
	if err := c.decodeJSONResponse(resp, &v); err != nil {
		return err
	}
	return nil
}

func (c *Client) BatchDownload(urls []string, outPath string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(urls))
	for _, URL := range urls {
		wg.Add(1)
		go func(URL string) {
			defer wg.Done()
			slug, _, err := c.ID(URL)
			if err != nil {
				errChan <- err
				return
			}
			pathname, err := c.PathName(TypeClip, slug, outPath)
			if err != nil {
				errChan <- err
				return
			}
			if err := c.DownloadClip(slug, pathname); err != nil {
				errChan <- fmt.Errorf("failed to download clip from URL: %s , Error: \n%w", URL, err)
			}
		}(URL)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		return err
	}
	return nil
}

func (c *Client) IsChannelLive(channelName string) (bool, error) {
	u := fmt.Sprintf("https://decapi.me/twitch/uptime/%s", channelName)
	resp, err := http.Get(u)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	return !strings.Contains(string(b), "offline"), nil
}

func (c *Client) DownloadClip(slug, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create the outPath. Maybe the output that is provided is incorrect: %s", err)
	}
	defer out.Close()
	creds, err := c.GetClipCreds(slug)
	if err != nil {
		return err
	}
	stream, err := c.ClipStream(creds)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, stream)
	if err != nil {
		return fmt.Errorf("failed to write the stream into outPath: %s", err)
	}
	return nil
}

// func (c *Client) DownloadVideo(name, id, quality string, start, end time.Duration) error {
// 	token, sig, err := c.GetVideoCredentials(id)
// 	if err != nil {
// 		return err
// 	}
// 	m3u8, err := c.GetVODMasterM3u8(token, sig, id)
// 	if err != nil {
// 		return err
// 	}
// 	serialized := string(m3u8)
// 	fmt.Println("SErialized", serialized)
// 	urls := c.GetMediaPlaylists(serialized)
// 	u := file.ConstructURL(urls, quality)
// 	if err := c.DownloadVOD(u, name, start, end); err != nil {
// 		return err
// 	}
// 	return nil
// }

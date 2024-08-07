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

	"github.com/schollz/progressbar/v3"
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
func (c *Client) MediaName(id string, vType VideoType) (string, error) {
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
// func (c *Client) PathName(vType VideoType, id, output string) (string, error) {
// 	name, err := c.extractNameFromID(vType, id)
// 	if err != nil {
// 		return "", err
// 	}
// 	return name, nil
// }

func (c *Client) ID(URL string) (string, VideoType, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse the URL: %s", err)
	}
	if !strings.Contains(u.Hostname(), "twitch.tv") {
		return "", 0, fmt.Errorf("the hostname of the URL does not contain twitch.tv")
	}
	// handle live stream
	s := strings.Split(u.Path, "/")
	if len(s) == 2 {
		return s[1], TypeLivestream, nil
	}
	// handle clip
	if strings.Contains(u.Path, "/clip/") {
		_, id := path.Split(u.Path)
		return id, TypeClip, nil
	}
	// handle vod
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

func (c *Client) NewGetRequest(URL string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()
	return req, nil
}

func (c *Client) decodeJSONResponse(resp *http.Response, p interface{}) error {
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return err
	}
	return nil
}

func (c *Client) sendGqlLoadAndDecode(body *strings.Reader, v any) error {
	req, err := http.NewRequest(http.MethodPost, c.gqlURL, body)
	if err != nil {
		return fmt.Errorf("failed to create request to get the access token: %s", err)
	}
	req.Header.Set("Client-Id", c.gqlClientID)
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	if err := c.decodeJSONResponse(resp, &v); err != nil {
		return err
	}
	return nil
}

func IsChannelLive(channelName string) (bool, error) {
	u := fmt.Sprintf("https://decapi.me/twitch/uptime/%s", channelName)
	resp, err := http.Get(u)
	if err != nil {
		return false, fmt.Errorf("failed getting the response from URL: %s. \nError: %s", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("channel %s does not exist?", channelName)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed reading the response Body. \nError: %s", err)
	}
	if strings.HasPrefix(string(b), "[Error from Twitch API]") {
		return false, fmt.Errorf("unexpected error")
	}
	return !strings.Contains(string(b), "offline"), nil
}

func (c *Client) downloadSegment(req *http.Request, destPath string, bar *progressbar.ProgressBar) error {
	f, err := os.OpenFile(destPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get the response from: %s", req.URL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response status: %s", resp.Status)
	}
	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return err
	}
	return nil
}

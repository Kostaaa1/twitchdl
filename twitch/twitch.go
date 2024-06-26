package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	utils "github.com/Kostaaa1/twitchdl/utils/file"
)

const (
	gqlURL      = "https://gql.twitch.tv/gql"
	gqlClientID = "kimne78kx3ncx6brgo4mv6wki5h1ko"
	usherURL    = "https://usher.ttvnw.net"
	helixURL    = "https://api.twitch.tv/helix"
)

type ScriptResponse struct {
	Message string
	Status  string
}

type Client struct {
	client      *http.Client
	clientID    string
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

func (c *Client) PathName(vType VideoType, id, output string) (string, error) {
	name, err := c.Name(vType, id)
	if err != nil {
		return "", err
	}
	name = utils.CreateVideo(output, name)
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

func New(client *http.Client, clientID string) Client {
	return Client{
		client:      client,
		clientID:    clientID,
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

func (c *Client) fetch(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request failed: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching m3u8 failed: %w", err)
	}
	defer resp.Body.Close()

	m3bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %w", err)
	}
	return m3bytes, nil
}

func (c *Client) AppendToFile(filename string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

func (c *Client) RequestInterceptor(w http.ResponseWriter, r *http.Request, outpath, addr string) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal("reading body failed")
	}
	var res ScriptResponse
	if err := json.Unmarshal(b, &res); err != nil {
		log.Fatalf("failed to unmarshal: %s", err)
	}
	go func() {
		tsBytes, err := c.fetch(res.Message)
		if err != nil {
			fmt.Println("fetching tsBytes error in goroutine")
			return
		}
		log.Printf("size of segment: %d bytes\n", len(tsBytes))
		if len(tsBytes) > 100000 {
			if err := c.AppendToFile(outpath, tsBytes); err != nil {
				http.Error(w, "Failed to append segment", http.StatusInternalServerError)
				log.Println("failed to append segment:", err)
				return
			}
		}
	}()
}

func (c *Client) RecordLivetream(outPath, streamURL string) {
	serverStarted := make(chan int)
	var port int
	go func(URL string) {
		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			log.Fatal(err)
		}
		defer listener.Close()

		port = listener.Addr().(*net.TCPAddr).Port
		http.HandleFunc("/segment", func(w http.ResponseWriter, r *http.Request) {
			c.RequestInterceptor(w, r, outPath, streamURL)
		})
		serverStarted <- port
		if err := http.Serve(listener, nil); err != nil {
			log.Fatal(err)
		}
	}(streamURL)

	<-serverStarted

	cmd := exec.Command("node", "./scripts/index.js", streamURL, strconv.Itoa(port))
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func (c *Client) BatchDownload(urls []string, outPath string) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(urls))

	for _, URL := range urls {
		wg.Add(1)
		go func(URL string) {
			defer wg.Done()
			id, _, err := c.ID(URL)
			if err != nil {
				errChan <- err
				return
			}
			name, err := c.PathName(TypeClip, id, outPath)
			if err != nil {
				errChan <- err
				return
			}
			if err := c.DownloadClip(name, id); err != nil {
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
	s := string(b)
	return !strings.Contains(s, "offline"), nil
}

func (c *Client) DownloadClip(filepath, slug string) error {

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

func (c *Client) DownloadVideo(name, id, quality string, start, end time.Duration) error {
	token, sig, err := c.GetVideoCredentials(id)
	if err != nil {
		return err
	}
	m3u8, err := c.GetMasterM3u8(token, sig, id)
	if err != nil {
		return err
	}
	serialized := string(m3u8)
	urls := c.GetMediaPlaylists(serialized)
	u := utils.ConstructURL(urls, quality)
	if err := c.DownloadVOD(u, name, start, end); err != nil {
		return err
	}
	return nil
}

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
	"time"

	"github.com/Kostaaa1/twitchdl/internal/config"
	"github.com/Kostaaa1/twitchdl/internal/file"
	"github.com/Kostaaa1/twitchdl/types"
)

type Client struct {
	config      types.JsonConfig
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

func (c *Client) MediaName(id string, vType VideoType) (string, error) {
	switch vType {
	case TypeClip:
		clip, err := c.ClipMetadata(id)
		if err != nil {
			return "", err
		}
		id = fmt.Sprintf("%s - %s", clip.Broadcaster.DisplayName, clip.Title)
	case TypeVOD:
		vod, err := c.VideoMetadata(id)
		if err != nil {
			fmt.Println("error", err)
			return "", err
		}
		id = fmt.Sprintf("%s - %s", vod.Owner.Login, vod.Title)
	}
	return id, nil
}

func (c *Client) ID(URL string) (string, VideoType, error) {
	parsedURL, err := url.Parse(URL)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse the URL: %s", err)
	}
	if !strings.Contains(parsedURL.Hostname(), "twitch.tv") {
		return "", 0, fmt.Errorf("the hostname of the URL does not contain twitch.tv")
	}
	s := strings.Split(parsedURL.Path, "/")

	if strings.Contains(parsedURL.Host, "clips.twitch.tv") || strings.Contains(parsedURL.Path, "/clip/") {
		_, id := path.Split(parsedURL.Path)
		return id, TypeClip, nil
	}
	if strings.Contains(parsedURL.Path, "/videos/") {
		_, id := path.Split(parsedURL.Path)
		return id, TypeVOD, nil
	}
	return s[1], TypeLivestream, nil
}

func New() *Client {
	cfg, err := config.Get()
	if err != nil {
		panic(err)
	}
	return &Client{
		client:      http.DefaultClient,
		config:      *cfg,
		gqlURL:      "https://gql.twitch.tv/gql",
		gqlClientID: "kimne78kx3ncx6brgo4mv6wki5h1ko",
		usherURL:    "https://usher.ttvnw.net",
		helixURL:    "https://api.twitch.tv/helix",
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

func (c *Client) GetToken() string {
	return fmt.Sprintf("Bearer %s", c.config.Creds.AccessToken)
}

type progressWriter struct {
	writer     io.Writer
	progressCh chan<- types.ProgresbarChanData
	slug       string
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err == nil {
		select {
		case pw.progressCh <- types.ProgresbarChanData{
			Text:  pw.slug,
			Bytes: int64(n),
		}:
		default:
		}
	}
	return n, err
}

func (c *Client) BatchDownload(urls []string, quality, destpath string, start, end time.Duration, progressCh chan types.ProgresbarChanData) error {
	cLimit := len(urls)
	var wg sync.WaitGroup
	sem := make(chan struct{}, cLimit)
	for _, URL := range urls {
		wg.Add(1)
		go func(URL string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := c.Downloader(URL, destpath, quality, start, end, progressCh); err != nil {
				fmt.Println(err)
				return
			}
		}(URL)
	}
	wg.Wait()
	close(progressCh)
	return nil
}

func (api *Client) Downloader(URL, destPath, quality string, start, end time.Duration, progressCh chan types.ProgresbarChanData) error {
	slug, vType, err := api.ID(URL)
	if err != nil {
		return err
	}
	mediaName, _ := api.MediaName(slug, vType)
	finalDest := file.NewPathname(destPath, mediaName)

	switch vType {
	case TypeVOD:
		if err := api.DownloadVideo(URL, slug, quality, finalDest, start, end, progressCh); err != nil {
			fmt.Println(err)
			return err
		}
	case TypeClip:
		if err := api.DownloadClip(slug, quality, finalDest); err != nil {
			return err
		}
	case TypeLivestream:
		if err := api.RecordStream(slug, quality, destPath); err != nil {
			return err
		}
	}
	// send that download is completed to the chan
	progressCh <- types.ProgresbarChanData{
		Text:   URL,
		IsDone: true,
	}
	return nil
}

func (c *Client) downloadSegment(destPath string, req *http.Request, pw *progressWriter) error {
	f, err := os.OpenFile(destPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := c.client.Do(req)
	if err != nil {
		fmt.Println("FAILED TO GET THE RESPONSEJ", err)
		return fmt.Errorf("failed to get the response from: %s", req.URL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	_, err = io.Copy(pw, resp.Body)
	if err != nil {
		fmt.Println("Failed to copy to pw: ", err)
		return err
	}
	return nil
}

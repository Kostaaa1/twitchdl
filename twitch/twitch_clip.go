package twitch

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

type VideoQualities []struct {
	Typename  string  `json:"__typename"`
	FrameRate float64 `json:"frameRate"`
	Quality   string  `json:"quality"`
	SourceURL string  `json:"sourceURL"`
}

type ClipCredentials struct {
	Typename            string `json:"__typename"`
	ID                  string `json:"id"`
	PlaybackAccessToken struct {
		Typename  string `json:"__typename"`
		Signature string `json:"signature"`
		Value     string `json:"value"`
	} `json:"playbackAccessToken"`
	VideoQualities VideoQualities `json:"videoQualities"`
	// VideoQualities []struct {
	// 	Typename  string  `json:"__typename"`
	// 	FrameRate float64 `json:"frameRate"`
	// 	Quality   string  `json:"quality"`
	// 	SourceURL string  `json:"sourceURL"`
	// } `json:"videoQualities"`
}

func (c *Client) extractSourceURL(videoQualities VideoQualities, quality string) string {
	if quality == "best" {
		return videoQualities[0].SourceURL
	}
	if quality == "worst" {
		return videoQualities[len(videoQualities)-1].SourceURL
	}
	for _, q := range videoQualities {
		if strings.HasPrefix(q.Quality, quality) {
			return q.SourceURL
		}
	}
	return quality
}

func (c *Client) GetClipData(slug string) (ClipCredentials, error) {
	gqlPayload := `{
        "operationName": "VideoAccessToken_Clip",
        "variables": {
            "slug": "%s"
        },
        "extensions": {
            "persistedQuery": {
                "version": 1,
                "sha256Hash": "36b89d2507fce29e5ca551df756d27c1cfe079e2609642b4390aa4c35796eb11"
            }
        }
    }`
	type payload struct {
		Data struct {
			Clip ClipCredentials `json:"clip"`
		} `json:"data"`
	}
	var p payload
	body := strings.NewReader(fmt.Sprintf(gqlPayload, slug))
	if err := c.sendGqlLoadAndDecode(body, &p); err != nil {
		return ClipCredentials{}, err
	}
	return p.Data.Clip, nil
}

func (c *Client) GetClipUsherURL(sourceURL, sig, token string) string {
	URL := fmt.Sprintf("%s?sig=%s&token=%s", sourceURL, url.QueryEscape(sig), url.QueryEscape(token))
	return URL
}

func (c *Client) DownloadClip(slug, quality, destPath string) error {
	clip, err := c.GetClipData(slug)
	if err != nil {
		return err
	}
	sourceURL := c.extractSourceURL(clip.VideoQualities, quality)
	usherURL := c.GetClipUsherURL(sourceURL, clip.PlaybackAccessToken.Signature, clip.PlaybackAccessToken.Value)

	req, err := http.NewRequest(http.MethodGet, usherURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create the new request for stream: %s", err)
	}
	req.Header.Set("Client-Id", c.gqlClientID)
	resp, err := c.client.Do(req)
	if err != nil {
		log.Fatal("SKRT: ", err)
	}
	defer resp.Body.Close()

	bar := progressbar.DefaultBytes(-1, "Recording:")
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create the outPath. Maybe the output that is provided is incorrect: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write the stream into outPath: %s", err)
	}
	return nil
}

func (c *Client) BatchDownload(urls []string, quality, destPath string) error {
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
			if err := c.DownloadClip(slug, quality, destPath); err != nil {
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

type ClipMetadata struct {
	Typename    string `json:"__typename"`
	Broadcaster struct {
		Typename    string `json:"__typename"`
		DisplayName string `json:"displayName"`
		ID          string `json:"id"`
	} `json:"broadcaster"`
	CreatedAt       time.Time `json:"createdAt"`
	DurationSeconds int       `json:"durationSeconds"`
	Game            struct {
		Typename string `json:"__typename"`
		ID       string `json:"id"`
		Name     string `json:"name"`
	} `json:"game"`
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (c *Client) ClipMetadata(slug string) (ClipMetadata, error) {
	gqlPayload := `{
        "operationName": "ComscoreStreamingQuery",
        "variables": {
            "channel": "",
            "clipSlug": "%s",
            "isClip": true,
            "isLive": false,
            "isVodOrCollection": false,
            "vodID": ""
        },
        "extensions": {
            "persistedQuery": {
                "version": 1,
                "sha256Hash": "e1edae8122517d013405f237ffcc124515dc6ded82480a88daef69c83b53ac01"
            }
        }
    }`
	type payload struct {
		Data struct {
			Clip ClipMetadata `json:"clip"`
		} `json:"data"`
	}
	var p payload
	body := strings.NewReader(fmt.Sprintf(gqlPayload, slug))
	if err := c.sendGqlLoadAndDecode(body, &p); err != nil {
		return ClipMetadata{}, err
	}
	return p.Data.Clip, nil
}

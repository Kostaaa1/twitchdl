package twitch

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
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

func (c *Client) GetClipUsherURL(slug, quality string) (string, error) {
	clip, err := c.GetClipData(slug)
	if err != nil {
		return "", err
	}
	sourceURL := c.extractSourceURL(clip.VideoQualities, quality)
	URL := fmt.Sprintf("%s?sig=%s&token=%s", sourceURL, url.QueryEscape(clip.PlaybackAccessToken.Signature), url.QueryEscape(clip.PlaybackAccessToken.Value))
	return URL, nil
}

func (c *Client) DownloadClip(slug, quality, destPath string) error {
	usherURL, err := c.GetClipUsherURL(slug, quality)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, usherURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create the new request for stream: %s", err)
	}
	req.Header.Set("Client-Id", c.gqlClientID)
	// if err := c.downloadSegment(req, destPath); err != nil {
	// 	return err
	// }
	return nil
}

// func (c *Client) BatchDownload(urls []string, quality, destpath string) error {
// 	cLimit := 4
// 	var wg sync.WaitGroup
// 	errChan := make(chan error, len(urls))
// 	sem := make(chan struct{}, cLimit)
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()
// 	for _, URL := range urls {
// 		wg.Add(1)
// 		go func(URL string) {
// 			defer wg.Done()
// 			select {
// 			case sem <- struct{}{}:
// 			case <-ctx.Done():
// 				return
// 			}
// 			defer func() { <-sem }()
// 			slug, vtype, err := c.ID(URL)
// 			if err != nil {
// 				errChan <- err
// 				return
// 			}
// 			if vtype != 0 {
// 				errChan <- fmt.Errorf("non clip URL is provided: %s. Batches work only with clips", URL)
// 				cancel()
// 				return
// 			}
// 			if err := c.DownloadClip(slug, quality, destpath); err != nil {
// 				errChan <- err
// 				cancel()
// 				return
// 			}
// 		}(URL)
// 	}
// 	wg.Wait()
// 	close(errChan)
// 	if len(errChan) > 0 {
// 		for err := range errChan {
// 			fmt.Println(err)
// 		}
// 	}
// 	return nil
// }

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

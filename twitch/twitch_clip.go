package twitch

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ClipCredentials struct {
	Typename            string `json:"__typename"`
	ID                  string `json:"id"`
	PlaybackAccessToken struct {
		Typename  string `json:"__typename"`
		Signature string `json:"signature"`
		Value     string `json:"value"`
	} `json:"playbackAccessToken"`
	VideoQualities []struct {
		Typename  string  `json:"__typename"`
		FrameRate float64 `json:"frameRate"`
		Quality   string  `json:"quality"`
		SourceURL string  `json:"sourceURL"`
	} `json:"videoQualities"`
}

func (c *Client) GetClipCreds(slug string) (ClipCredentials, error) {
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

	body := strings.NewReader(fmt.Sprintf(gqlPayload, slug))
	req, err := http.NewRequest(http.MethodPost, c.gqlURL, body)
	if err != nil {
		return ClipCredentials{}, fmt.Errorf("failed to create request to get the access token: %s", err)
	}
	req.Header.Set("Client-Id", c.gqlClientID)

	type payload struct {
		Data struct {
			Clip ClipCredentials `json:"clip"`
		} `json:"data"`
	}
	var p payload

	resp, err := c.do(req)
	if err != nil {
		return ClipCredentials{}, err
	}
	if err := c.decodeJSONResponse(resp, &p); err != nil {
		return ClipCredentials{}, err
	}

	return p.Data.Clip, nil
}

func (c *Client) ClipStream(clip ClipCredentials) (io.ReadCloser, error) {
	URL := fmt.Sprintf("%s?sig=%s&token=%s", clip.VideoQualities[0].SourceURL, url.QueryEscape(clip.PlaybackAccessToken.Signature), url.QueryEscape(clip.PlaybackAccessToken.Value))
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create the new request for stream: %s", err)
	}
	req.Header.Set("Client-Id", c.gqlClientID)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream response: %s", err)
	}
	return resp.Body, nil
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

	body := strings.NewReader(fmt.Sprintf(gqlPayload, slug))
	req, err := http.NewRequest(http.MethodPost, c.gqlURL, body)
	if err != nil {
		return ClipMetadata{}, fmt.Errorf("failed to create request to get the clip data: %s", err)
	}
	req.Header.Set("Client-Id", c.gqlClientID)

	type payload struct {
		Data struct {
			Clip ClipMetadata `json:"clip"`
		} `json:"data"`
	}
	var p payload
	resp, err := c.do(req)
	if err != nil {
		return ClipMetadata{}, err
	}
	if err := c.decodeJSONResponse(resp, &p); err != nil {
		return ClipMetadata{}, err
	}
	return p.Data.Clip, nil
}

package twitch

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type VODCredentials struct {
	Typename  string `json:"__typename"`
	Signature string `json:"signature"`
	Value     string `json:"value"`
}

func (c *Client) GetVideoCredentials(id string) (string, string, error) {
	gqlPayload := `{
        "operationName": "PlaybackAccessToken_Template",
        "query": "query PlaybackAccessToken_Template($login: String!, $isLive: Boolean!, $vodID: ID!, $isVod: Boolean!, $playerType: String!) {  streamPlaybackAccessToken(channelName: $login, params: {platform: \"web\", playerBackend: \"mediaplayer\", playerType: $playerType}) @include(if: $isLive) {    value    signature   authorization { isForbidden forbiddenReasonCode }   __typename  }  videoPlaybackAccessToken(id: $vodID, params: {platform: \"web\", playerBackend: \"mediaplayer\", playerType: $playerType}) @include(if: $isVod) {    value    signature   __typename  }}",
        "variables": {
            "isLive": false,
            "login": "",
            "isVod": true,
            "vodID": "%s",
            "playerType": "site"
        }
    }`

	body := strings.NewReader(fmt.Sprintf(gqlPayload, id))
	req, err := http.NewRequest(http.MethodPost, c.gqlURL, body)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request to get the access token: %s", err)
	}
	req.Header.Set("Client-Id", c.gqlClientID)

	type payload struct {
		Data struct {
			VideoPlaybackAccessToken VODCredentials `json:"videoPlaybackAccessToken"`
		} `json:"data"`
	}
	var p payload
	resp, err := c.do(req)
	if err != nil {
		return "", "", err
	}
	if err := c.decodeJSONResponse(resp, &p); err != nil {
		return "", "", err
	}
	return p.Data.VideoPlaybackAccessToken.Value, p.Data.VideoPlaybackAccessToken.Signature, nil
}

func (c *Client) GetMediaPlaylists(serializedM3u8 string) []string {
	lines := strings.Split(serializedM3u8, "\n")
	var u []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF") {
			u = append(u, lines[i+1])
		}
	}
	return u
}

func (c *Client) GetMasterM3u8(token, sig, id string) ([]byte, error) {
	u := fmt.Sprintf("%s/vod/%s?nauth=%s&nauthsig=%s&allow_audio_only=true&allow_source=true",
		c.usherURL, id, token, sig)
	resp, err := c.client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if s := resp.StatusCode; s < 200 || s >= 300 {
		return nil, fmt.Errorf("unsupported status code (%v) for url: %s", s, u)
	}
	return io.ReadAll(resp.Body)
}

type VideoMetadata struct {
	Typename  string    `json:"__typename"`
	CreatedAt time.Time `json:"createdAt"`
	Game      struct {
		Typename    string `json:"__typename"`
		DisplayName string `json:"displayName"`
		ID          string `json:"id"`
	} `json:"game"`
	ID    string `json:"id"`
	Owner struct {
		Typename string `json:"__typename"`
		ID       string `json:"id"`
		Login    string `json:"login"`
	} `json:"owner"`
	Title string `json:"title"`
}

func (c *Client) VideoMetadata(id string) (VideoMetadata, error) {
	gqlPayload := `{
		"operationName": "NielsenContentMetadata",
		"variables": {
			"isCollectionContent": false,
			"isLiveContent": false,
			"isVODContent": true,
			"collectionID": "",
			"login": "",
			"vodID": "%s"
		},
		"extensions": {
			"persistedQuery": {
				"version": 1,
				"sha256Hash": "2dbf505ee929438369e68e72319d1106bb3c142e295332fac157c90638968586"
			}
		}
	}`

	body := strings.NewReader(fmt.Sprintf(gqlPayload, id))
	req, err := http.NewRequest(http.MethodPost, c.gqlURL, body)
	if err != nil {
		return VideoMetadata{}, fmt.Errorf("failed to create request to get the access token: %s", err)
	}
	req.Header.Set("Client-Id", c.gqlClientID)

	type payload struct {
		Data struct {
			Video VideoMetadata `json:"video"`
		} `json:"data"`
	}
	var p payload
	resp, err := c.do(req)
	if err != nil {
		return VideoMetadata{}, err
	}
	if err := c.decodeJSONResponse(resp, &p); err != nil {
		return VideoMetadata{}, err
	}
	return p.Data.Video, nil
}

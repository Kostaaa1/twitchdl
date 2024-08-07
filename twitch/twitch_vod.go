package twitch

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	file "github.com/Kostaaa1/twitchdl/utils"
	"github.com/schollz/progressbar/v3"
)

type VideoCredResponse struct {
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
	type payload struct {
		Data struct {
			VideoPlaybackAccessToken VideoCredResponse `json:"videoPlaybackAccessToken"`
		} `json:"data"`
	}
	var p payload

	body := strings.NewReader(fmt.Sprintf(gqlPayload, id))
	if err := c.sendGqlLoadAndDecode(body, &p); err != nil {
		return "", "", err
	}
	return p.Data.VideoPlaybackAccessToken.Value, p.Data.VideoPlaybackAccessToken.Signature, nil
}

func (c *Client) GetVODMasterM3u8(token, sig, id string) ([]byte, error) {
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

	type payload struct {
		Data struct {
			Video VideoMetadata `json:"video"`
		} `json:"data"`
	}
	var p payload

	body := strings.NewReader(fmt.Sprintf(gqlPayload, id))
	if err := c.sendGqlLoadAndDecode(body, &p); err != nil {
		return VideoMetadata{}, err
	}
	return p.Data.Video, nil
}

func (c *Client) DownloadVideo(dstPath, id, quality string, start, end time.Duration) error {
	token, sig, err := c.GetVideoCredentials(id)
	if err != nil {
		return err
	}
	m3u8, err := c.GetVODMasterM3u8(token, sig, id)
	if err != nil {
		return err
	}
	serialized := string(m3u8)

	urls := c.GetMediaPlaylists(serialized)
	u := file.ConstructURL(urls, quality)
	if err := c.DownloadVOD(u, dstPath, start, end); err != nil {
		return err
	}
	return nil
}

func (c *Client) DownloadVOD(URL, filePath string, start, end time.Duration) error {
	m3u8, err := c.GetMediaPlaylist(fmt.Sprintf("%sindex-dvr.m3u8", URL))
	if err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	lines := strings.Split(string(m3u8), "\n")
	var segmentDuration float64 = 10
	segStart := 0

	s := int(start.Seconds()/segmentDuration) * 2
	e := int(end.Seconds()/segmentDuration) * 2

	var segmentLines []string
	if e == 0 {
		segmentLines = lines[segStart:][s:]
	} else {
		segmentLines = lines[segStart:][s:e]
	}

	bar := progressbar.DefaultBytes(-1, "Downloading:")
	for _, tsFile := range segmentLines {
		if strings.HasSuffix(tsFile, ".ts") {
			chunkURL := fmt.Sprintf("%s%s", URL, tsFile)
			req, err := http.NewRequest(http.MethodGet, chunkURL, nil)
			if err != nil {
				fmt.Println("failed to create request for: ", chunkURL)
				break
			}
			if err := c.downloadSegment(req, filePath, bar); err != nil {
				fmt.Println("failed to download segment: ", chunkURL, "Error: ", err)
			}
			bar.Add(1)
		}
	}
	return nil
}

func (c *Client) GetMediaPlaylist(URL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	m3u8, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// m3u8, err := c.readResponseBody(resp)
	// if err != nil {
	// 	return nil, err
	// }
	return m3u8, nil
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

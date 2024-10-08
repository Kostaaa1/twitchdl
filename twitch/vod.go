package twitch

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/m3u8"
)

func (c *Client) DownloadVOD(unit MediaUnit) error {
	////////////// move it to single function ///////////////
	token, sig, err := c.GetVideoCredentials(unit.Slug)
	if err != nil {
		return err
	}
	master, err := c.GetVODMasterM3u8(token, sig, unit.Slug)
	if err != nil {
		return err
	}
	mediaList, err := master.GetVariantPlaylistByQuality(unit.Quality)
	if err != nil {
		return err
	}
	playlist, err := c.fetch(mediaList.URL)
	if err != nil {
		return err
	}
	///////////////////////////////////////////////////////

	var segmentDuration float64 = 10
	s := int(unit.Start.Seconds()/segmentDuration) * 2
	e := int(unit.End.Seconds()/segmentDuration) * 2

	var segmentLines []string
	lines := strings.Split(string(playlist), "\n")[8:]
	if e == 0 {
		segmentLines = lines[s:]
	} else {
		segmentLines = lines[s:e]
	}

	for _, tsFile := range segmentLines {
		if strings.HasSuffix(tsFile, ".ts") {
			lastIndex := strings.LastIndex(mediaList.URL, "/")
			chunkURL := fmt.Sprintf("%s/%s", mediaList.URL[:lastIndex], tsFile)

			req, err := http.NewRequest(http.MethodGet, chunkURL, nil)
			if err != nil {
				fmt.Println("failed to create request for: ", chunkURL)
				return err
			}

			if err := c.downloadSegment(unit.DestPath, req, unit.pw); err != nil {
				fmt.Println("failed to download segment: ", chunkURL, "Error: ", err)
				return err
			}
		}
	}

	return nil
}

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

func (c *Client) GetVODMasterM3u8(token, sig, id string) (*m3u8.MasterPlaylist, error) {
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
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return m3u8.New(b), nil

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

// func (c *Client) FetchMediaPlaylist(playlistURL string) ([]byte, error) {
// 	req, err := http.NewRequest(http.MethodGet, playlistURL, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	resp, err := c.do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()
// 	m3u8, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return m3u8, nil
// }

func (c *Client) GetMediaPlaylists(master []byte) []string {
	lines := strings.Split(string(master), "\n")
	var u []string
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF") {
			u = append(u, lines[i+1])
		}
	}
	return u
}

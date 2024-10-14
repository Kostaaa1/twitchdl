package twitch

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/internal/m3u8"
)

func processSegments(mediaPlaylist []byte, start, end time.Duration) []string {
	var segmentDuration float64 = 10
	s := int(start.Seconds()/segmentDuration) * 2
	e := int(end.Seconds()/segmentDuration) * 2

	var segments []string
	lines := strings.Split(string(mediaPlaylist), "\n")[8:]
	if e == 0 {
		segments = lines[s:]
	} else {
		segments = lines[s:e]
	}
	return segments
}

func (c *Client) DownloadVOD(unit MediaUnit) error {
	////////////// move it to single function ///////////////
	master, status, err := c.GetVODMasterM3u8(unit.Slug)
	if err != nil && status != http.StatusForbidden {
		return err
	}
	var vodPlaylistURL string
	if status == http.StatusForbidden {
		subUrl, err := c.getSubVODPlaylistURL(unit.Slug, unit.Quality)
		if err != nil {
			return err
		}
		vodPlaylistURL = subUrl
	} else {
		variantList, err := master.GetVariantPlaylistByQuality(unit.Quality)
		if err != nil {
			return err
		}
		vodPlaylistURL = variantList.URL
	}
	mediaPlaylist, err := c.fetch(vodPlaylistURL)
	if err != nil {
		return err
	}
	////////////////////////////////////////////////////////////

	segments := processSegments(mediaPlaylist, unit.Start, unit.End)

	for _, tsFile := range segments {
		if strings.HasSuffix(tsFile, ".ts") {
			lastIndex := strings.LastIndex(vodPlaylistURL, "/")
			chunkURL := fmt.Sprintf("%s/%s", vodPlaylistURL[:lastIndex], tsFile)

			req, err := http.NewRequest(http.MethodGet, chunkURL, nil)
			if err != nil {
				fmt.Println("failed to create request for: ", chunkURL)
				return err
			}

			if err := c.downloadSegment(req, unit.pw); err != nil {
				fmt.Println("failed to downloamediaList.URLd segment: ", chunkURL, "Error: ", err)
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

func (c *Client) getVideoCredentials(id string) (string, string, error) {
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

	type payload struct {
		Data struct {
			VideoPlaybackAccessToken VideoCredResponse `json:"videoPlaybackAccessToken"`
		} `json:"data"`
	}
	var p payload
	if err := c.sendGqlLoadAndDecode(body, &p); err != nil {
		return "", "", err
	}

	return p.Data.VideoPlaybackAccessToken.Value, p.Data.VideoPlaybackAccessToken.Signature, nil
}

func (c *Client) GetVODMasterM3u8(slug string) (*m3u8.MasterPlaylist, int, error) {
	token, sig, err := c.getVideoCredentials(slug)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	u := fmt.Sprintf("%s/vod/%s?nauth=%s&nauthsig=%s&allow_audio_only=true&allow_source=true", c.usherURL, slug, token, sig)

	b, code, err := c.fetchWithCode(u)
	if err != nil {
		return nil, code, err
	}

	return m3u8.New(b), code, nil
}

type SubVODResponse struct {
	Data struct {
		Video struct {
			BroadcastType string    `json:"broadcastType"`
			CreatedAt     time.Time `json:"createdAt"`
			Owner         struct {
				Login string `json:"login"`
			} `json:"owner"`
			SeekPreviewsURL string `json:"seekPreviewsURL"`
		} `json:"video"`
	} `json:"data"`
	Extensions struct {
		DurationMilliseconds int    `json:"durationMilliseconds"`
		RequestID            string `json:"requestID"`
	} `json:"extensions"`
}

func (c *Client) getSubVODPlaylistURL(slug, quality string) (string, error) {
	gqlPayload := `{
				"query": "query { video(id: \"%s\") { broadcastType, createdAt, seekPreviewsURL, owner { login } } }"
			}`
	body := strings.NewReader(fmt.Sprintf(gqlPayload, slug))

	var p SubVODResponse
	if err := c.sendGqlLoadAndDecode(body, &p); err != nil {
		return "", err
	}

	previewURL, err := url.Parse(p.Data.Video.SeekPreviewsURL)
	if err != nil {
		return "", err
	}

	paths := strings.Split(previewURL.Path, "/")
	var vodId string
	for i, p := range paths {
		if p == "storyboards" {
			vodId = paths[i-1]
		}
	}

	// [not tested] Only old uploaded VOD works with this method now
	// days_difference - difference between current date and p.Data.Video.CreatedAt
	// if broadcastType == "upload" && days_difference > 7 {
	// url = fmt.Sprintf(`https://${domain}/${channelData.login}/${vodId}/${vodSpecialID}/${resKey}/index-dvr.m3u8`, previewURL.Host, p.Data.Video.Owner.Login, slug, vodId, resolution)
	// }

	resolution := getResolution(quality)
	broadcastType := strings.ToLower(p.Data.Video.BroadcastType)
	var url string
	if broadcastType == "highlight" {
		url = fmt.Sprintf(`https://%s/%s/%s/highlight-%s.m3u8`, previewURL.Host, vodId, resolution, slug)
	} else if broadcastType != "upload" {
		url = fmt.Sprintf(`https://%s/%s/%s/index-dvr.m3u8`, previewURL.Host, vodId, resolution)
	}
	return url, nil
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

func (c *Client) videoMetadata(id string) (VideoMetadata, error) {
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

package twitch

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/m3u8"
)

func (c *Client) GetLivestreamCreds(id string) (string, string, error) {
	gqlPl := `{
		"operationName": "PlaybackAccessToken_Template",
		"query": "query PlaybackAccessToken_Template($login: String!, $isLive: Boolean!, $vodID: ID!, $isVod: Boolean!, $playerType: String!) {  streamPlaybackAccessToken(channelName: $login, params: {platform: \"web\", playerBackend: \"mediaplayer\", playerType: $playerType}) @include(if: $isLive) {    value    signature   authorization { isForbidden forbiddenReasonCode }   __typename  }  videoPlaybackAccessToken(id: $vodID, params: {platform: \"web\", playerBackend: \"mediaplayer\", playerType: $playerType}) @include(if: $isVod) {    value    signature   __typename  }}",
		"variables": {
			"isLive": true,
			"login": "%s",
			"isVod": false,
			"vodID": "",
			"playerType": "site"
		}
	}`
	type payload struct {
		Data struct {
			VideoPlaybackAccessToken VideoCredResponse `json:"streamPlaybackAccessToken"`
		} `json:"data"`
	}
	var data payload
	body := strings.NewReader(fmt.Sprintf(gqlPl, id))
	if err := c.sendGqlLoadAndDecode(body, &data); err != nil {
		return "", "", err
	}
	return data.Data.VideoPlaybackAccessToken.Value, data.Data.VideoPlaybackAccessToken.Signature, nil
}

func (c *Client) GetStreamMasterPlaylist(channel string) (string, error) {
	isLive, err := c.IsChannelLive(channel)
	if err != nil {
		return "", err
	}
	if !isLive {
		return "", fmt.Errorf("the channel %s is not live currently", channel)
	}

	tok, sig, err := c.GetLivestreamCreds(channel)
	if err != nil {
		return "", fmt.Errorf("failed to get livestream credentials: %w", err)
	}
	u := fmt.Sprintf("%s/api/channel/hls/%s.m3u8?token=%s&sig=%s&allow_audio_only=true&allow_source=true",
		c.usherURL, channel, tok, sig)

	resp, err := c.client.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if s := resp.StatusCode; s < 200 || s >= 300 {
		return "", fmt.Errorf("unsupported status code (%v) for url: %s", s, u)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Client) GetStreamMediaPlaylist(channel, quality string) (*m3u8.List, error) {
	master, err := c.GetStreamMasterPlaylist(channel)
	if err != nil {
		return nil, err
	}
	parsed := m3u8.Parse(master)
	mediaList, err := parsed.GetMediaPlaylist(quality)
	if err != nil {
		return nil, fmt.Errorf("failed to get media playlist: %w", err)
	}
	return &mediaList, nil
}

func isAdRunning(segments []string) int {
	for i := len(segments) - 1; i > 0; i-- {
		if segments[i] == "#EXT-X-DISCONTINUITY" {
			return i
		}
	}
	return 0
}

// Checks if the channel is live, gets the stream media playlist, creates the file,
func (c *Client) RecordStream(id, quality, outpath string) error {
	isLive, err := c.IsChannelLive(id)
	if err != nil {
		return err
	}
	if !isLive {
		return fmt.Errorf("the channel %s is currently offline", id)
	}

	mediaList, err := c.GetStreamMediaPlaylist(id, quality)
	if err != nil {
		return fmt.Errorf("failed to get media playlist: %w", err)
	}

	destPath := fmt.Sprintf("%s/%s - livestream-%s.mp4", outpath, id, time.Now().Format("2006-01-02-15-04-05"))
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	progressTicker := time.NewTicker(200 * time.Second)
	defer progressTicker.Stop()

	isAdFound := false

	for {
		select {
		case <-ticker.C:
			b, err := c.Fetch(mediaList.URL)
			if err != nil {
				return fmt.Errorf("failed to fetch playlist: %w", err)
			}
			segments := strings.Split(string(b), "\n")
			discontinuityID := isAdRunning(segments)
			if discontinuityID == 0 {
				isAdFound = false
				// tsURL := segments[len(segments)-2]
				// req, err := http.NewRequest(http.MethodGet, tsURL, nil)
				// if err != nil {
				// 	log.Println("failed to initiate the request")
				// }
				// if err := c.downloadSegment(req, destPath); err != nil {
				// 	log.Printf("failed to download and write segment: %v", err)
				// }
			} else {
				if !isAdFound {
					fmt.Printf("\n[Please be patient] found twitch AD at %d position, this can take a while...\n", discontinuityID)
					isAdFound = true
				}
			}
			// case <-progressTicker.C:
			// 	bar.Add(0)
		}
	}
}

func (c *Client) OpenStreamInMediaPlayer(channel string) error {
	media, err := c.GetStreamMediaPlaylist(channel, "best")
	if err != nil {
		return err
	}
	cmd := exec.Command("vlc", media.URL)
	if err := cmd.Run(); err != nil {
		fmt.Println("EXECUTION ERROR")
		return err
	}
	return nil
}

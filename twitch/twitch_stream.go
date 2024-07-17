package twitch

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/m3u8"
	"github.com/schollz/progressbar/v3"
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

	if err := c.sendGraphqlLoadAndDecode(body, &data); err != nil {
		return "", "", err
	}
	return data.Data.VideoPlaybackAccessToken.Value, data.Data.VideoPlaybackAccessToken.Signature, nil
}

func (c *Client) GetMasterStreamPlaylistURL(token, sig, id string) (string, error) {
	u := fmt.Sprintf("%s/api/channel/hls/%s.m3u8?token=%s&sig=%s&allow_audio_only=true&allow_source=true",
		c.usherURL, id, token, sig)
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

func (c *Client) GetMasterStreamPlaylist(id string) (string, error) {
	tok, sig, err := c.GetLivestreamCreds(id)
	if err != nil {
		return "", fmt.Errorf("failed to get livestream credentials: %w", err)
	}
	master, err := c.GetMasterStreamPlaylistURL(tok, sig, id)
	if err != nil {
		return "", fmt.Errorf("failed to get master stream playlist: %w", err)
	}
	return master, nil
}

func (c *Client) StartRecording(streamURL, quality, outpath string) error {
	id, _, err := c.ID(streamURL)
	if err != nil {
		log.Fatal(err)
	}
	isLive, err := c.IsChannelLive(id)
	if err != nil {
		return err
	}

	if isLive {
		newPath := fmt.Sprintf("%s/%s - livestream-%s.mp4", outpath, id, time.Now().Format("2006-01-02-15-04-05"))
		c.recordLivestream(id, streamURL, quality, newPath)
	} else {
		return fmt.Errorf("the channel %s is not live. In order to record the livestream, the channel needs to be live", id)
	}
	return nil
}

func (c *Client) recordLivestream(id, streamURL, quality, outPath string) error {
	master, err := c.GetMasterStreamPlaylist(id)
	if err != nil {
		return err
	}
	fmt.Println(master)
	parsed := m3u8.Parse(master)
	fmt.Println(parsed.GetQualities())
	masterList, err := parsed.GetMasterList(quality)
	if err != nil {
		return fmt.Errorf("failed to get media playlist: %w", err)
	}

	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	bar := progressbar.DefaultBytes(-1, "Recording:")
	defer bar.Close()

	for {
		select {
		case <-ticker.C:
			b, err := c.Fetch(masterList.URL)
			isAdActive := bytes.Contains(b, []byte("#EXT-X-DISCONTINUITY"))
			if err != nil {
				return fmt.Errorf("failed to fetch playlist: %w", err)
			}
			segments := strings.Split(string(b), "\n")
			if !isAdActive {
				// tsURL := segments[len(segments)-2]
				tsURL := getTSSegment(segments)
				if err := c.downloadAndWriteSegment(tsURL, outPath, bar); err != nil {
					log.Printf("failed to download and write segment: %v", err)
				}
			}
		}
	}
}

func (c *Client) downloadAndWriteSegment(tsURL, outPath string, bar *progressbar.ProgressBar) error {
	resp, err := c.client.Get(tsURL)
	if err != nil {
		return fmt.Errorf("failed to get segment: %w", err)
	}
	defer resp.Body.Close()

	f, err := os.OpenFile(outPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write ts content to file: %w", err)
	}
	return nil
}

func getTSSegment(segments []string) string {
	for _, seg := range segments {
		if strings.HasPrefix(seg, "http") && strings.HasSuffix(seg, ".ts") {
			return seg
		}
	}
	return ""
}

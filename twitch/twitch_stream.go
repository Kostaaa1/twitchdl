package twitch

import (
	"bytes"
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
		return "", fmt.Errorf("%s is offline", channel)
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
func (c *Client) RecordStream(slug, quality, destpath string, pw *progressWriter) error {
	isLive, err := c.IsChannelLive(slug)
	if err != nil {
		return err
	}
	if !isLive {
		return fmt.Errorf("%s is offline", slug)
	}

	mediaList, err := c.GetStreamMediaPlaylist(slug, quality)
	if err != nil {
		return fmt.Errorf("failed to get media playlist: %w", err)
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	tickCount := 0
	var halfBytes *bytes.Reader

	for {
		select {
		case <-ticker.C:
			tickCount++

			f, err := os.OpenFile(destpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			defer f.Close()

			if tickCount%2 != 0 {
				b, err := c.Fetch(mediaList.URL)
				if err != nil {
					return fmt.Errorf("failed to fetch playlist: %w", err)
				}
				segments := strings.Split(string(b), "\n")
				tsURL := segments[len(segments)-2]

				bodyBytes, err := c.Fetch(tsURL)
				if err != nil {
					return err
				}

				half := len(bodyBytes) / 2
				halfBytes = bytes.NewReader(bodyBytes[half:])
				if _, err := io.Copy(pw, bytes.NewReader(bodyBytes[:half])); err != nil {
					return err
				}
			}

			if tickCount%2 == 0 && halfBytes.Len() > 0 {
				if _, err := io.Copy(pw, halfBytes); err != nil {
					return err
				}
				halfBytes.Reset([]byte{})
			}
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

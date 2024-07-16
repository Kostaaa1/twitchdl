package twitch

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
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

	if err := c.sendGraphqlLoadAndDecode(body, &data); err != nil {
		return "", "", err
	}
	return data.Data.VideoPlaybackAccessToken.Value, data.Data.VideoPlaybackAccessToken.Signature, nil
}

func (c *Client) GetMasterStreamPlaylist(token, sig, id string) (string, error) {
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

func (c *Client) recordLivestream(id, streamURL, quality, path string) error {
	tok, sig, err := c.GetLivestreamCreds(id)
	if err != nil {
		return fmt.Errorf("failed to get livestream credentials: %w", err)
	}
	master, err := c.GetMasterStreamPlaylist(tok, sig, id)
	if err != nil {
		return fmt.Errorf("failed to get master stream playlist: %w", err)
	}
	parsed := m3u8.Parse(master)
	masterList, err := parsed.GetMediaPlaylist(quality)
	if err != nil {
		return fmt.Errorf("failed to get media playlist: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.downloadAndWriteSegment(masterList.URL, f); err != nil {
				log.Printf("failed to download and write segment: %v", err)
			}
		}
	}
}

func (c *Client) downloadAndWriteSegment(masterListURL string, f *os.File) error {
	bytes, err := c.Fetch(masterListURL)
	if err != nil {
		return fmt.Errorf("failed to fetch playlist: %w", err)
	}

	mediaList, _ := strconv.Unquote(string(bytes))
	segments := strings.Split(mediaList, "\n")
	if len(segments) < 2 {
		return fmt.Errorf("unexpected playlist format: %s", mediaList)
	}

	tsURL := segments[len(segments)-2]
	resp, err := c.client.Get(tsURL)
	if err != nil {
		return fmt.Errorf("failed to get segment: %w", err)
	}
	defer resp.Body.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write ts content to file: %w", err)
	}
	fmt.Printf("%d bytes written to file\n", n)
	return nil
}

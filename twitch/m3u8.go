package twitch

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
)

type MediaSegment struct {
	Number   int
	Duration time.Duration
	URL      string
}

type MediaPlaylist struct {
	BaseURL       string
	SegDuration   string
	TotalSeconds  string
	SegLength     int
	SegStartIndex int
}

func (c *Client) DownloadVOD(u, filePath string, start, end time.Duration) error {
	playlist := MediaPlaylist{}
	tsFileURL := fmt.Sprintf("%sindex-dvr.m3u8", u)

	resp, err := c.client.Get(tsFileURL)
	if err != nil {
		return fmt.Errorf("download failed. failed at getting the ts files from playlist: %s", err)
	}
	defer resp.Body.Close()

	if s := resp.StatusCode; s < 200 || s >= 300 {
		return fmt.Errorf("status code: %v. Download failed. failed at getting the ts files from playlist: %s", s, err)
	}

	m3u8, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body bytes: %s", err)
	}

	lines := strings.Split(string(m3u8), "\n")
	var segmentDuration float64 = 10
	segStart := -1

	s := int(start.Seconds()/segmentDuration) * 2
	e := int(end.Seconds()/segmentDuration) * 2

	playlist.SegLength = len(lines)
	for i := 0; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "#EXT-X-TARGETDURATION") {
			playlist.SegDuration = strings.Split(lines[i], "#EXT-X-TARGETDURATION:")[1]
		}
		if strings.HasPrefix(lines[i], "#EXT-X-TWITCH-TOTAL-SECS:") {
			playlist.TotalSeconds = strings.Split(lines[i], "#EXT-X-TWITCH-TOTAL-SECS:")[1]
		}
		if segStart == -1 && strings.HasPrefix(lines[i], "#EXTINF:") {
			playlist.SegStartIndex = i
			segStart = i
			break
		}
	}
	if segStart == -1 {
		return fmt.Errorf("no segments found in the m3u8 playlist")
	}

	var segmentLines []string
	if e == 0 {
		segmentLines = lines[segStart:][s:]
	} else {
		segmentLines = lines[segStart:][s:e]
	}

	bar := progressbar.NewOptions(len(segmentLines)/2,
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetDescription("Downloading: "),
	)

	segments := []string{}
	for i := 0; i < len(segmentLines); i++ {
		if strings.HasPrefix(segmentLines[i], "#EXTINF") {
			chunkURL := fmt.Sprintf("%s%s", u, segmentLines[i+1])
			fname, err := c.downloadSegment(chunkURL)
			if err != nil {
				fmt.Printf("Failed to download segment %s: %s", segmentLines[i+1], fname)
				break
			}
			segments = append(segments, fname)
			bar.Add(1)
		}
	}
	if err := writeSegments(filePath, segments); err != nil {
		fmt.Println("\nFailed to concat the segments: ", err)
	}
	return nil
}

func (c *Client) downloadSegment(tsFile string) (string, error) {
	resp, err := c.client.Get(tsFile)
	if err != nil {
		err := fmt.Errorf("failed to get the tsFile content: %s", err)
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	file, err := os.CreateTemp("", "segment-*.ts")
	if err != nil {
		err := fmt.Errorf("failed to create temp: %s", err)
		fmt.Println(err)
		return "", err
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		err := fmt.Errorf("failed to io.Copy into file: %s", err)
		fmt.Println(err)
		return "", err
	}
	return file.Name(), nil
}

func writeSegments(filePath string, segments []string) error {
	cmd := exec.Command("ffmpeg", "-f", "mpegts", "-i", "concat:"+strings.Join(segments, "|"), "-c", "copy", filePath)
	err := cmd.Run()
	if err != nil {
		fmt.Printf("failed to concatenate segments: %s", err)
		return fmt.Errorf("failed to concatenate segments: %s", err)
	}
	return nil
}

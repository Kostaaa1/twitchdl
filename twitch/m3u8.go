package twitch

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/utils"
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

func (c *Client) GetMediaPlaylist(URL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	m3u8, err := c.readResponseBody(resp)
	if err != nil {
		return nil, err
	}
	return m3u8, nil
}

func (c *Client) DownloadVOD(URL, filePath string, start, end time.Duration) error {
	playlist := MediaPlaylist{}
	tsFileURL := fmt.Sprintf("%sindex-dvr.m3u8", URL)

	m3u8, err := c.GetMediaPlaylist(tsFileURL)
	if err != nil {
		return err
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

	for _, tsFile := range segmentLines {
		if strings.HasSuffix(tsFile, ".ts") {
			chunkURL := fmt.Sprintf("%s%s", URL, tsFile)
			if err := c.downloadAndAppend(filePath, chunkURL); err != nil {
				fmt.Println("FAILED TO DOWNLOAD AND APPEND: ", chunkURL)
			}
			bar.Add(1)
		}
	}

	return nil
}

func (c *Client) downloadAndAppend(outPath, tsFile string) error {
	resp, err := c.client.Get(tsFile)
	if err != nil {
		err := fmt.Errorf("failed to get the tsFile content: %s", err)
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed ot read segment respbody: %w", err)
	}
	if err = utils.AppendToFile(outPath, b); err != nil {
		return fmt.Errorf("failed to append segment bytes to file: %w", err)
	}
	return nil
}

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/internal/config"
	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/types"
	"github.com/Kostaaa1/twitchdl/view/chat"
	"github.com/Kostaaa1/twitchdl/view/components"
)

type Prompt struct {
	Url      string        `json:"url"`
	Quality  string        `json:"quality"`
	Start    time.Duration `json:"start"`
	End      time.Duration `json:"end"`
	DestPath string        `json:"destPath"`
}

func (p *Prompt) UnmarshalJSON(b []byte) error {
	type Alias Prompt
	aux := &struct {
		Start string `json:"start"`
		End   string `json:"end"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}

	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	var err error
	if aux.Start != "" {
		p.Start, err = time.ParseDuration(aux.Start)
		if err != nil {
			return err
		}
	}

	if aux.End != "" {
		p.Start, err = time.ParseDuration(aux.Start)
		if err != nil {
			return err
		}
	}
	return nil
}

func (prompt *Prompt) processInput(tw *twitch.Client) ([]twitch.MediaUnit, chan types.ProgresbarChanData) {
	if prompt.Url == "" {
		log.Fatalf("Input was not provided.")
	}
	var units []twitch.MediaUnit
	var progressCh chan types.ProgresbarChanData

	_, err := os.Stat(prompt.Url)

	if os.IsNotExist(err) {
		urls := strings.Split(prompt.Url, ",")
		progressCh = make(chan types.ProgresbarChanData, len(urls))

		for _, url := range urls {
			unit, err := tw.NewMediaUnit(url, prompt.Quality, prompt.DestPath, prompt.Start, prompt.End, progressCh)
			if err != nil {
				continue
			}
			units = append(units, unit)
		}
	} else {
		content, err := os.ReadFile(prompt.Url)
		if err != nil {
			panic(err)
		}
		var body []Prompt
		if err := json.Unmarshal(content, &body); err != nil {
			panic(err)
		}

		progressCh = make(chan types.ProgresbarChanData, len(body))
		for _, b := range body {
			unit, err := tw.NewMediaUnit(b.Url, b.Quality, b.DestPath, b.Start, b.End, progressCh)
			if err != nil {
				fmt.Println(err)
				continue
			}
			units = append(units, unit)
		}
	}
	return units, progressCh
}

func main() {
	jsonCfg, err := config.Get()
	if err != nil {
		panic(err)
	}

	var prompt Prompt
	flag.StringVar(&prompt.Url, "input", "", "The URL of the clip to download. You can download multiple clips as well by seperating them by comma (no spaces in between). Exapmle: -url https://www.twitch.tv/{...}")
	flag.StringVar(&prompt.Quality, "quality", "best", "[best 1080 720 480 360 160 worst]. Example: -quality 1080p (optional)")
	flag.DurationVar(&prompt.Start, "start", time.Duration(0), "The start of the VOD subset. It only works with VODs and it needs to be in this format: '1h30m0s' (optional)")
	flag.DurationVar(&prompt.End, "end", time.Duration(0), "The end of the VOD subset. It only works with VODs and it needs to be in this format: '1h33m0s' (optional)")
	flag.StringVar(&prompt.DestPath, "output", jsonCfg.Paths.OutputPath, "Path to the downloaded video.")
	flag.Parse()

	tw := twitch.New()
	if prompt.Url == "" {
		if len(os.Args) > 1 {
			prompt.Url = os.Args[1]
		} else {
			// root.Open(twitch, jsonCfg)
			chat.Open(tw, jsonCfg)
			return
		}
	}

	units, progressCh := prompt.processInput(tw)
	tw.SetProgressChannel(progressCh)

	go func() {
		paths := make([]string, len(units))
		for i, u := range units {
			paths[i] = u.DestPath
		}
		components.Spinner(paths, progressCh)
	}()

	if len(units) > 1 {
		if err := tw.BatchDownload(units); err != nil {
			panic(err)
		}
	} else {
		if err := tw.Downloader(units[0]); err != nil {
			panic(err)
		}
	}

	close(progressCh)

	time.Sleep(500 * time.Millisecond)
	fmt.Println("Finished downloading")
	fmt.Printf("\033[?25h")
}

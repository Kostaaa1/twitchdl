package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/internal/config"
	"github.com/Kostaaa1/twitchdl/spinner"
	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/types"
	"github.com/Kostaaa1/twitchdl/view/root"
)

type Config struct {
	inputURL   string
	quality    string
	start, end time.Duration
	output     string
}

func main() {
	jsonCfg, err := config.Get()
	if err != nil {
		panic(err)
	}

	var cfg Config
	flag.StringVar(&cfg.inputURL, "input", "", "The URL of the clip to download. You can download multiple clips as well by seperating them by comma (no spaces in between). Exapmle: -url https://www.twitch.tv/{...}")
	flag.StringVar(&cfg.quality, "quality", "best", "[best 1080 720 480 360 160 worst]. Example: -quality 1080p (optional)")
	flag.DurationVar(&cfg.start, "start", time.Duration(0), "The start of the VOD subset. It only works with VODs and it needs to be in this format: '1h30m0s' (optional)")
	flag.DurationVar(&cfg.end, "end", time.Duration(0), "The end of the VOD subset. It only works with VODs and it needs to be in this format: '1h33m0s' (optional)")
	flag.StringVar(&cfg.output, "output", jsonCfg.Paths.OutputPath, "Path to the downloaded video.")
	flag.Parse()

	twitch := twitch.New()
	if cfg.inputURL == "" {
		if len(os.Args) > 1 {
			cfg.inputURL = os.Args[1]
		} else {
			root.Open(twitch, jsonCfg)
			return
		}
	}
	urls := strings.Split(cfg.inputURL, ",")
	progressCh := make(chan types.ProgresbarChanData, len(urls))

	go func() {
		spinner.Open(urls, progressCh)
	}()

	if len(urls) > 1 {
		if err := twitch.BatchDownload(urls, cfg.quality, cfg.output, cfg.start, cfg.end, progressCh); err != nil {
			panic(err)
		}
	} else {
		if err := twitch.Downloader(urls[0], cfg.output, cfg.quality, cfg.start, cfg.end, progressCh); err != nil {
			panic(err)
		}
	}
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("Finished downloading")
}

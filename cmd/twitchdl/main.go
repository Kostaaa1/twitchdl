package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Kostaaa1/twitchdl/internal/config"
	"github.com/Kostaaa1/twitchdl/internal/prompt"
	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/view/components"
)

func main() {
	jsonCfg, err := config.Get()
	if err != nil {
		panic(err)
	}

	var prompt prompt.Prompt
	flag.StringVar(&prompt.Input, "input", "", "The URL of the clip to download. You can download multiple clips as well by seperating them by comma (no spaces in between). Exapmle: -url https://www.twitch.tv/{...}")
	flag.StringVar(&prompt.Quality, "quality", "", "[best 1080 720 480 360 160 worst]. Example: -quality 1080p (optional)")
	flag.DurationVar(&prompt.Start, "start", time.Duration(0), "The start of the VOD subset. It only works with VODs and it needs to be in this format: '1h30m0s' (optional)")
	flag.DurationVar(&prompt.End, "end", time.Duration(0), "The end of the VOD subset. It only works with VODs and it needs to be in this format: '1h33m0s' (optional)")
	flag.StringVar(&prompt.Output, "output", jsonCfg.Paths.OutputPath, "Path to the downloaded video.")
	flag.Parse()

	tw := twitch.New()
	if prompt.Input == "" {
		if len(os.Args) > 1 {
			prompt.Input = os.Args[1]
		}
	}

	units, progressCh := prompt.ProcessInput(tw)
	tw.SetProgressChannel(progressCh)

	if jsonCfg.ShowDownloadSpinner {
		go func() {
			paths := make([]string, len(units))
			for i, u := range units {
				paths[i] = u.DestPath
			}
			components.Spinner(paths, progressCh)
		}()
	}

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
	fmt.Println("Finished downloading 👍👍👍")
	fmt.Printf("\033[?25h")
}

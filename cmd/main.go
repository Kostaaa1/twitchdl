package main

import (
	"flag"
	"log"
	"os"
	"time"
)

type Config struct {
	inputURL   string
	quality    string
	start, end time.Duration
	output     string
	overwrite  bool
}

type Client struct {
	logger *log.Logger
	config *Config
}

func createNewClient() *Client {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	return &Client{
		logger: logger,
		config: &Config{},
	}
}

func main() {
	client := createNewClient()
	outpath := "/mnt/c/Users/kosta/OneDrive/Desktop/imgs/Clips"
	// paths, err := db.GetBucketValues(client.db)
	// if err != nil {
	// 	client.logger.Fatal(err)
	// }

	var cfg Config
	flag.StringVar(&cfg.inputURL, "input", "", "The URL of the clip to download. You can download multiple clips as well by seperating them by comma (no spaces in between). Exapmle: -url https://www.twitch.tv/{...}, https://twitch.tv/{...}")
	flag.StringVar(&cfg.quality, "quality", "best", "[best 1080 720 480 360 160 worst]. Example: -quality 1080p (optional)")
	flag.DurationVar(&cfg.start, "start", time.Duration(0), "The start of the VOD subset. It only works with VODs and it needs to be in this format: '1h30m0s' (optional)")
	flag.DurationVar(&cfg.end, "end", time.Duration(0), "The end of the VOD subset. It only works with VODs and it needs to be in this format: '1h33m0s' (optional)")
	flag.BoolVar(&cfg.overwrite, "overwrite", false, "Overwrite the database paths with provided paths.")
	flag.StringVar(&cfg.output, "output", outpath, "Path to the downloaded video.")
	flag.Parse()
	if cfg.inputURL == "" {
		if len(os.Args) > 1 {
			cfg.inputURL = os.Args[1]
		} else {
			client.logger.Panic("You need to provide the twitch URL bozo.")
		}
	}

	// bar := progressbar.DefaultBytes(-1, "Downloading:")
	// api := twitch.New()
	// if err := api.Downloader(cfg.output, cfg.inputURL, cfg.quality, cfg.start, cfg.end, bar); err != nil {
	// 	log.Fatal(err)
	// }
}

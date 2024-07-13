package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/utils"
)

type Config struct {
	recordURL   string
	inputURL    string
	quality     string
	start, end  time.Duration
	overwrite   bool
	jsoncfg     utils.JSONConfig
	isLogConfig bool
}

func main() {
	var cfg Config
	defaultCfg, _ := utils.ReadConfig()

	flag.StringVar(&cfg.inputURL, "url", "", "The URL of the clip to download. You can download multiple clips as well by seperating them by comma (no spaces in between). Exapmle: -url https://www.twitch.tv/{...},https://twitch.tv/{...}")
	flag.StringVar(&cfg.quality, "quality", "", "[1080p 720p 480p 360p]. Example: -quality 1080p (optional)")
	flag.StringVar(&cfg.recordURL, "record", "", "Record the livestream. Example: -record https:twitch.tv/pokimane")
	flag.DurationVar(&cfg.start, "start", time.Duration(0), "The start of the VOD subset. It only works with VODs and it needs to be in this format: '1h30m0s' (optional)")
	flag.DurationVar(&cfg.end, "end", time.Duration(0), "The end of the VOD subset. It only works with VODs and it needs to be in this format: '1h33m0s' (optional)")
	flag.BoolVar(&cfg.overwrite, "overwrite", false, "This value will override the jsonconfig with provided values.")
	flag.BoolVar(&cfg.isLogConfig, "logconfig", false, "Print the config.json")
	flag.StringVar(&cfg.jsoncfg.Outpath, "outpath", defaultCfg.Outpath, "Path to the downloaded video.")
	flag.StringVar(&cfg.jsoncfg.JSPath, "jspath", defaultCfg.JSPath, "Path to the puppeteer js file.")

	flag.Parse()

	if cfg.isLogConfig {
		utils.PrintConfig(cfg.jsoncfg)
		return
	}
	err := utils.ArePathsValid(cfg.jsoncfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.overwrite {
		if utils.IsNotEmpty(cfg.jsoncfg) {
			utils.OverwriteConfig(cfg.jsoncfg)
		} else {
			log.Println("There are no overwritable fields for JSON config.")
		}
	}

	if err := cfg.run(); err != nil {
		log.Fatal(err)
	}
}

func (cfg *Config) run() error {
	api := twitch.New(http.DefaultClient)
	jsoncfg := cfg.jsoncfg
	if cfg.recordURL != "" {
		if err := api.StartRecording(cfg.recordURL, jsoncfg.Outpath, jsoncfg.JSPath); err != nil {
			return err
		}
		return nil
	}
	batch := strings.Split(cfg.inputURL, ",")
	if len(batch) > 1 {
		if err := api.BatchDownload(batch, jsoncfg.Outpath); err != nil {
			return err
		}
		return nil
	}
	id, vType, err := api.ID(cfg.inputURL)
	if err != nil {
		return err
	}
	name, err := api.PathName(vType, id, jsoncfg.Outpath)
	if err != nil {
		return err
	}
	if cfg.inputURL != "" {
		switch vType {
		case twitch.TypeClip:
			if err := api.DownloadClip(name, id); err != nil {
				return err
			}
		case twitch.TypeVOD:
			if cfg.quality != "" && !isValidQuality(cfg.quality) {
				return fmt.Errorf("the quality that you provided is not supported")
			}
			if err := api.DownloadVideo(name, id, cfg.quality, cfg.start, cfg.end); err != nil {
				return err
			}
		}
	}
	return nil
}

func isValidQuality(q string) bool {
	switch q {
	case "1080p", "720p", "480p", "360p", "160p":
		return true
	default:
		return false
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Kostaaa1/twitchdl/twitch"
	utils "github.com/Kostaaa1/twitchdl/utils/file"
)

var (
	inputURLs, recordURL, inputURL, quality string
	defaultOut, clientID, output            string
	start, end                              time.Duration
	logPath                                 bool
)

const (
	Quality1080 string = "1080p"
	Quality720  string = "720p"
	Quality480  string = "480p"
	Quality360  string = "360p"
	Quality160  string = "160p"
)

func getOutPath() string {
	if logPath {
		err := utils.LogDefaultPath()
		if err != nil {
			log.Fatal(err)
		}
	}
	if defaultOut != "" {
		var v utils.Config
		v.OutPath = defaultOut
		utils.SetConfig("./config.json", v)
	}
	var out string
	if output != "" {
		out = output
	} else {
		cfg, err := utils.ReadConfig()
		if err != nil {
			panic(err)
		}
		out = cfg.OutPath
	}
	return out
}

func main() {
	flag.StringVar(&inputURL, "url", "", "The URL of the clip to download.")
	flag.StringVar(&inputURLs, "batch", "", "Provide multiple clip URLs to download concurrently. Example: --batch url,url... (Separate them with comma)")
	flag.StringVar(&clientID, "client-id", "", "The Client ID to use the helix API.")
	flag.StringVar(&output, "output", "", "The path to the downloaded clips")
	flag.DurationVar(&start, "start", time.Duration(0), "The start of the VOD subset. It only works with VODs and it needs to be in this format: '1h30m0s' (optional)")
	flag.DurationVar(&end, "end", time.Duration(0), "The end of the VOD subset. It only works with VODs and it needs to be in this format: '1h33m0s' (optional)")
	flag.StringVar(&quality, "quality", "", "[1080p 720p 480p 360p]. Example: --quality 1080p (optional)")
	flag.StringVar(&defaultOut, "set-default-out", "", "Provide the default path where to store the downloaded videos. Example: --set-default-out ./home/user/downloads")
	flag.BoolVar(&logPath, "default-out", false, "Your default output path.")
	flag.StringVar(&recordURL, "record", "", "Listen to requests and download them. Example: --listen https:twitch.tv/pokimane")
	flag.Parse()

	out := getOutPath()
	if err := run(out); err != nil {
		log.Fatal(err)
	}

}

func run(outPath string) error {
	api := twitch.New(http.DefaultClient, clientID)
	if recordURL != "" {
		newPath := fmt.Sprintf("output-%s.mp4", time.Now().Format("2006-01-02-15-04-05"))
		f, err := os.Create(newPath)
		if err != nil {
			return err
		}
		api.RecorcdStream(f.Name(), recordURL)
		return nil
	}

	if inputURLs != "" {
		if err := api.BatchDownload(inputURLs, outPath); err != nil {
			return err
		}
		return nil
	}

	id, vType, err := api.ID(inputURL)
	if err != nil {
		return err
	}

	name, err := api.PathName(vType, id, outPath)
	if err != nil {
		return err
	}

	if inputURL != "" {
		switch vType {
		case twitch.TypeClip:
			if err := api.DownloadClip(name, id); err != nil {
				return err
			}
		case twitch.TypeVOD:
			if quality != "" && !isValidQuality(quality) {
				return fmt.Errorf("the quality that you provided is not supported")
			}
			if err := api.DownloadVideo(name, id, quality, start, end); err != nil {
				return err
			}
		}
	}
	return nil
}

func isValidQuality(q string) bool {
	switch q {
	case Quality1080, Quality720, Quality480, Quality360, Quality160:
		return true
	default:
		return false
	}
}

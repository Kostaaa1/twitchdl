package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/db"
	"github.com/Kostaaa1/twitchdl/twitch"
	file "github.com/Kostaaa1/twitchdl/utils"
	"github.com/boltdb/bolt"
)

type Config struct {
	inputURL   string
	quality    string
	start, end time.Duration
	// db paths, maybe remove?
	output     string
	printPaths bool
	jspath     string
	overwrite  bool
}

type Client struct {
	logger *log.Logger
	db     *bolt.DB
	config *Config
}

func createNewClient() *Client {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	newDb, err := db.SetupDB()
	if err != nil {
		logger.Fatal(err)
	}
	return &Client{
		db:     newDb,
		logger: logger,
		config: &Config{},
	}
}

func main() {
	client := createNewClient()
	paths, err := db.GetBucketValues(client.db)
	if err != nil {
		client.logger.Fatal(err)
	}
	var cfg Config

	flag.StringVar(&cfg.inputURL, "input", os.Args[1], "The URL of the clip to download. You can download multiple clips as well by seperating them by comma (no spaces in between). Exapmle: -url https://www.twitch.tv/{...},https://twitch.tv/{...}")
	flag.StringVar(&cfg.quality, "quality", "best", "[1080p 720p 480p 360p]. Example: -quality 1080p (optional)")
	flag.DurationVar(&cfg.start, "start", time.Duration(0), "The start of the VOD subset. It only works with VODs and it needs to be in this format: '1h30m0s' (optional)")
	flag.DurationVar(&cfg.end, "end", time.Duration(0), "The end of the VOD subset. It only works with VODs and it needs to be in this format: '1h33m0s' (optional)")
	flag.BoolVar(&cfg.overwrite, "overwrite", false, "Overwrite the database paths with provided paths.")
	flag.BoolVar(&cfg.printPaths, "printPaths", false, "Print the provided paths. If printPaths=true, other options wont execute.")
	flag.StringVar(&cfg.output, "output", paths.Outpath, "Path to the downloaded video.")
	flag.StringVar(&cfg.jspath, "jspath", paths.Jspath, "Path to the puppeteer js file.")

	flag.Parse()
	////////////////////
	if cfg.printPaths {
		db.PrintConfig(client.db)
		return
	}
	if cfg.overwrite {
		db.UpdateBucketValues(client.db, db.DBKeys{Outpath: cfg.output, Jspath: cfg.jspath})
	}
	////////////////////
	if !IsValidQuality(cfg.quality) {
		log.Printf("input quality (%s) is not supported", cfg.output)
		PrintQualities()
	}
	if err := cfg.run(); err != nil {
		log.Fatal(err)
	}
}

func (cfg *Config) run() error {
	out := cfg.output
	api := twitch.New(http.DefaultClient)

	id, vType, err := api.ID(cfg.inputURL)
	mediaName, _ := api.MediaName(id, vType)
	finalDest := file.CreatePathname(out, mediaName)

	if err != nil {
		return err
	}

	// if cfg.quality != ""
	// 	return fmt.Errorf("the quality that you provided is not supported")
	// }

	batch := strings.Split(cfg.inputURL, ",")
	if len(batch) > 1 {
		if err := api.BatchDownload(batch, cfg.quality, out); err != nil {
			return err
		}
		return nil
	}

	switch vType {
	case twitch.TypeVOD:
		// newPathName, err := api.PathName(vType, id, out)
		// if err != nil {
		// 	return err
		// }
		if err := api.DownloadVideo(finalDest, id, cfg.quality, cfg.start, cfg.end); err != nil {
			return err
		}
	case twitch.TypeClip:
		// newPathName, err := api.PathName(vType, id, out)
		// if err != nil {
		// 	return err
		// }
		if err := api.DownloadClip(id, cfg.quality, finalDest); err != nil {
			return err
		}
	case twitch.TypeLivestream:
		if err := api.StartRecording(id, cfg.quality, out); err != nil {
			return err
		}
	}
	return nil
}

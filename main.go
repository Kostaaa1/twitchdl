package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Kostaaa1/twitchdl/db"
	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/boltdb/bolt"
)

type Config struct {
	recordURL  string
	inputURL   string
	quality    string
	outpath    string
	jspath     string
	overwrite  bool
	printPaths bool
	start, end time.Duration
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

	flag.StringVar(&cfg.inputURL, "url", "", "The URL of the clip to download. You can download multiple clips as well by seperating them by comma (no spaces in between). Exapmle: -url https://www.twitch.tv/{...},https://twitch.tv/{...}")
	flag.StringVar(&cfg.quality, "quality", "best", "[1080p 720p 480p 360p]. Example: -quality 1080p (optional)")
	flag.StringVar(&cfg.recordURL, "record", "", "Record the livestream. Example: -record https:twitch.tv/pokimane")
	flag.DurationVar(&cfg.start, "start", time.Duration(0), "The start of the VOD subset. It only works with VODs and it needs to be in this format: '1h30m0s' (optional)")
	flag.DurationVar(&cfg.end, "end", time.Duration(0), "The end of the VOD subset. It only works with VODs and it needs to be in this format: '1h33m0s' (optional)")
	flag.BoolVar(&cfg.overwrite, "overwrite", false, "Overwrite the database paths with provided paths.")
	flag.BoolVar(&cfg.printPaths, "printPaths", false, "Print the provided paths. If printPaths=true, other options wont execute.")
	flag.StringVar(&cfg.outpath, "outpath", paths.Outpath, "Path to the downloaded video.")
	flag.StringVar(&cfg.jspath, "jspath", paths.Jspath, "Path to the puppeteer js file.")
	flag.Parse()

	if cfg.printPaths {
		db.PrintConfig(client.db)
		return
	}
	if cfg.overwrite {
		db.UpdateBucketValues(client.db, db.DBKeys{Outpath: cfg.outpath, Jspath: cfg.jspath})
	}
	if err := cfg.run(); err != nil {
		log.Fatal(err)
	}
}

func (cfg *Config) run() error {
	out := cfg.outpath
	api := twitch.New(http.DefaultClient)
	if cfg.recordURL != "" {
		if err := api.StartRecording(cfg.recordURL, cfg.quality, out); err != nil {
			return err
		}
		return nil
	}
	// batch := strings.Split(cfg.inputURL, ",")
	// if len(batch) > 1 {
	// 	if err := api.BatchDownload(batch, out); err != nil {
	// 		return err
	// 	}
	// 	return nil
	// }

	// id, vType, err := api.ID(cfg.inputURL)
	// if err != nil {
	// 	return err
	// }

	// name, err := api.PathName(vType, id, out)
	// fmt.Println("name", name)
	// if err != nil {
	// 	return err
	// }
	// if cfg.inputURL != "" {
	// 	switch vType {
	// 	case twitch.TypeClip:
	// 		if err := api.DownloadClip(name, id); err != nil {
	// 			return err
	// 		}
	// 	case twitch.TypeVOD:
	// 		if cfg.quality != "" && !IsValidQuality(cfg.quality) {
	// 			return fmt.Errorf("the quality that you provided is not supported")
	// 		}
	// 		if err := api.DownloadVideo(name, id, cfg.quality, cfg.start, cfg.end); err != nil {
	// 			return err
	// 		}
	// 	}
	// }
	return nil
}

func IsValidQuality(q string) bool {
	switch q {
	case "best", "1080p60", "720p60", "480p30", "360p30", "160p30":
		return true
	default:
		return false
	}
}

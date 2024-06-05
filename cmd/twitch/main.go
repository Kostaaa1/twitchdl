package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Kostaaa1/twitch-clip-downloader/twitch"
	utils "github.com/Kostaaa1/twitch-clip-downloader/utils/file"
)

var (
	inputURLs, inputURL, defaultOut, clientID, output, quality string
	start, end                                                 time.Duration
	logPath                                                    bool
)

const (
	Quality1080 string = "1080p"
	Quality720  string = "720p"
	Quality480  string = "480p"
	Quality360  string = "360p"
	Quality160  string = "160p"
)

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
	flag.Parse()

	if logPath {
		err := utils.LogDefaultPath()
		if err != nil {
			log.Fatal(err)
		}
		return
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

	if err := run(out); err != nil {
		log.Fatal(err)
	}
}

func batchDownload(api twitch.Client, URLs string) error {
	urls := strings.Split(URLs, ",")
	var wg sync.WaitGroup
	errChan := make(chan error, len(urls))

	for _, URL := range urls {
		wg.Add(1)
		go func(URL string) {
			defer wg.Done()
			id, _, err := api.ID(URL)
			if err != nil {
				errChan <- err
				return
			}

			name, err := api.PathName(twitch.TypeClip, id, output)
			if err != nil {
				errChan <- err
				return
			}

			if err := downloadClip(api, name, id); err != nil {
				errChan <- fmt.Errorf("failed to download clip from URL: %s , Error: \n%w", URL, err)
			}
		}(URL)
	}

	wg.Wait()
	close(errChan)
	for err := range errChan {
		return err
	}
	return nil
}

func downloadClip(api twitch.Client, filepath, slug string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create the outPath. Maybe the output that is provided is incorrect: %s", err)
	}
	defer out.Close()

	creds, err := api.GetClipCreds(slug)
	if err != nil {
		return err
	}
	stream, err := api.ClipStream(creds)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, stream)
	if err != nil {
		return fmt.Errorf("failed to write the stream into outPath: %s", err)
	}
	return nil
}

func fullURL(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}
	v, _ := path.Split(parsed.Path)
	fullURL := &url.URL{
		Scheme: "https",
		Host:   parsed.Host,
		Path:   v,
	}
	return fullURL.String()
}

func constructURL(urls []string, quality string) string {
	var u string
	if quality != "" {
		for _, x := range urls {
			if strings.Contains(x, quality) {
				u = fullURL(x)
			}
		}
	} else {
		u = fullURL(urls[0])
	}
	return u
}

func downloadVideo(api twitch.Client, name, id string) error {
	token, sig, err := api.GetVideoCredentials(id)
	if err != nil {
		return err
	}
	m3u8, err := api.GetMasterM3u8(token, sig, id)
	if err != nil {
		return err
	}
	serialized := string(m3u8)
	urls := api.GetMediaPlaylists(serialized)
	u := constructURL(urls, quality)

	if err := api.DownloadVOD(u, name, start, end); err != nil {
		return err
	}
	return nil
}

func run(outPath string) error {
	api := twitch.New(http.DefaultClient, clientID)
	if inputURLs != "" {
		if err := batchDownload(api, inputURLs); err != nil {
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
			if err := downloadClip(api, name, id); err != nil {
				return err
			}
		case twitch.TypeVOD:
			if quality != "" && !isValidQuality(quality) {
				return fmt.Errorf("the quality that you provided is not supported")
			}
			if err := downloadVideo(api, name, id); err != nil {
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

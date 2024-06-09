package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

type Config struct {
	OutPath string
}

func ReadFile(filePath string) ([]byte, error) {
	d, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the filePath: %s. error: %s", filePath, err)
	}
	defer d.Close()

	s, err := io.ReadAll(d)
	if err != nil {
		return nil, fmt.Errorf("failed to read the content: %s, error: %s", filePath, err)
	}

	return s, nil
}

func IsVideoQuality(segment string) bool {
	re := regexp.MustCompile(`^\d+p\d*$`)
	return re.MatchString(segment)
}

func ReadConfig() (Config, error) {
	bytes, err := ReadFile("./config.json")
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func LogDefaultPath() error {
	cfg, err := ReadConfig()
	if err != nil {
		return err
	}
	log.Print(cfg.OutPath)
	return nil
}

func SetConfig(filePath string, v Config) error {
	if FileExists(filePath) {
		data, err := json.MarshalIndent(v, "", " ")
		if err != nil {
			return err
		}
		err = os.WriteFile(filePath, data, 0644)
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("the filepath is incorrect: %s", filePath)
}

func CreateVideo(dir, filename string) string {
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	filename = re.ReplaceAllString(filename, "_")
	filePath := filepath.Join(dir, fmt.Sprintf("%s.mp4", filename))
	counter := 1
	for FileExists(filePath) {
		filePath = filepath.Join(dir, fmt.Sprintf("%s (%v).mp4", filename, counter))
		counter++
	}

	_, err := os.Create(filePath)
	if err != nil {
		log.Fatal("Failted to create new video")
	}
	return filePath
}

func AppendToFile(filePath string, data []byte) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open the files: %w", err)
	}
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

func getFullURL(u string) string {
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

func ConstructURL(urls []string, quality string) string {
	var u string
	if quality != "" {
		for _, x := range urls {
			if strings.Contains(x, quality) {
				u = getFullURL(x)
			}
		}
	} else {
		u = getFullURL(urls[0])
	}
	return u
}

package file

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func isFileReal(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
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

// func CreateVideo(dir, filename string) string {
// 	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
// 	filename = re.ReplaceAllString(filename, "_")
// 	filePath := filepath.Join(dir, fmt.Sprintf("%s.mp4", filename))
// 	counter := 1
// 	for isFileReal(filePath) {
// 		filePath = filepath.Join(dir, fmt.Sprintf("%s (%v).mp4", filename, counter))
// 		counter++
// 	}
// 	f, err := os.Create(filePath)
// 	if err != nil {
// 		log.Fatal("Failted to create new video")
// 	}
// 	defer func() {
// 		if err := f.Close(); err != nil {
// 			log.Print("failed to close the created video file: ", err)
// 		}
// 	}()
// 	return filePath
// }

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

func IncludeExecPath(path string) (string, error) {
	execPath, err := GetExecPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(execPath, path), nil
}

func GetExecPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execPath = filepath.Dir(execPath)
	return execPath, nil
}

func CreatePathname(dstPath, filename string) string {
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	filename = re.ReplaceAllString(filename, "_")
	filePath := filepath.Join(dstPath, fmt.Sprintf("%s.mp4", filename))
	counter := 1
	for isFileReal(filePath) {
		filePath = filepath.Join(dstPath, fmt.Sprintf("%s (%v).mp4", filename, counter))
		counter++
	}
	return filePath
}

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

func ConstructURL(urls []string, quality string) string {
	if quality == "best" {
		return getFullURL(urls[0])
	}
	if quality == "worst" {
		return getFullURL(urls[len(urls)-1])
	}
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

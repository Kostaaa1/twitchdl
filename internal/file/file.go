package file

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

func Exists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func NewPathname(dstPath, filename string) string {
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	filename = re.ReplaceAllString(filename, "_")
	filePath := filepath.Join(dstPath, fmt.Sprintf("%s.mp4", filename))
	counter := 1
	for Exists(filePath) {
		filePath = filepath.Join(dstPath, fmt.Sprintf("%s (%v).mp4", filename, counter))
		counter++
	}
	return filePath
}

// func IncludeExecPath(path string) (string, error) {
// 	execPath, err := GetExecPath()
// 	if err != nil {
// 		return "", err
// 	}
// 	return filepath.Join(execPath, path), nil
// }
// func GetExecPath() (string, error) {
// 	execPath, err := os.Executable()
// 	if err != nil {
// 		return "", err
// 	}
// 	execPath = filepath.Dir(execPath)
// 	return execPath, nil
// }

package utils

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// this is bad, optimize this
// func ConvertBytes(b float64) string {
// 	units := []string{"B", "KB", "MB", "GB", "TB"}
// 	i := 0
// 	for b >= 1024 && i < len(units)-1 {
// 		b /= 1024
// 		i++
// 	}
// 	return fmt.Sprintf("%.01f %s", b, units[i])
// }

func GetCurrentTimeFormatted() string {
	now := time.Now()
	timestamp := now.UnixNano() / int64(time.Millisecond)
	formattedTime := ParseTimestamp(fmt.Sprintf("%d", timestamp))
	return formattedTime
}

func ParseTimestamp(v string) string {
	timestamp, _ := strconv.ParseInt(v, 10, 64)
	seconds := timestamp / 1000
	t := time.Unix(seconds, 0)
	formatted := t.Format("03:04")
	return formatted
}

func Capitalize(v string) string {
	return strings.ToUpper(v[:1]) + v[1:]
}

func RemoveCursor() {
	fmt.Printf("\033[?25l")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Printf("\033[?25h")
		os.Exit(0)
	}()
}

func createServingID() string {
	w := strings.Split("0123456789abcdefghijklmnopqrstuvwxyz", "")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	id := ""
	for i := 0; i < 32; i++ {
		id += w[r.Intn(len(w))]
	}
	return id
}

func ConstructPathname(dstPath, ext string) (string, error) {
	info, err := os.Stat(dstPath)

	if os.IsNotExist(err) {
		if filepath.Ext(dstPath) != "" {
			dir := filepath.Dir(dstPath)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				return "", fmt.Errorf("directory does not exist: %s", dir)
			}
			return dstPath, nil
		}
		return "", fmt.Errorf("path does not exist: %s", dstPath)
	}

	if info.IsDir() {
		fileID := createServingID()
		filename := fmt.Sprintf("%s.%s", fileID, ext)
		newpath := filepath.Join(dstPath, filename)
		return newpath, nil
	}

	return "", fmt.Errorf("this path already exists %s: ", dstPath)
}

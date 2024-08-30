package utils

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// this is bad, optimize this
func ConvertBytes(b float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for b >= 1024 && i < len(units)-1 {
		b /= 1024
		i++
	}
	return fmt.Sprintf("%.01f %s", b, units[i])
}

// func ConvertBytes(b float64) string {
// 	if b < 1 {
// 		return fmt.Sprintf("%.01f B", b)
// 	}
// 	units := []string{"B", "KB", "MB", "GB", "TB"}
// 	exp := math.Min(float64(len(units)-1), math.Floor(math.Log2(b)/10))
// 	value := b / math.Pow(1024, exp)
// 	return fmt.Sprintf("%.01f %s", value, units[int(exp)])
// }

func CalcSpeed(b int64, elapsedTime float64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	values := make([]int64, len(units))
	values[0] = b
	for i := 1; i < len(units); i++ {
		values[i] = values[i-1] / 1024
	}
	for i := len(units) - 1; i >= 0; i-- {
		if values[i] > 0 {
			return fmt.Sprintf("%d %s", values[i], units[i])
		}
	}
	return ""
}

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

func Capitalize(v string) string {
	return strings.ToUpper(v[:1]) + v[1:]
}

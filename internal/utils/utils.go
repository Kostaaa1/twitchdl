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

// export function convertBytes(bytes: string) {
//   const b = Number(bytes);
//   const KB = b / 1024;
//   const MB = KB / 1024;
//   const GB = MB / 1024;
//   const TB = GB / 1024;

//   const result = {
//     bytes,
//     KB: KB.toFixed(2),
//     MB: MB.toFixed(2),
//     GB: GB.toFixed(2),
//     TB: TB.toFixed(2),
//   };

//   const index = Object.entries(result).findIndex((x) => x[1][0] === "0");
//   const output = Object.entries(result)[index - 1];
//   return output[1].split(".")[0] + output[0];
// }

func ConvertBytes(b int64) string {
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

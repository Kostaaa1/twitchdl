package main

import (
	"fmt"
	"strings"
)

// ////////////////////////
var supportedQualities = []string{"best", "1080p60", "720p60", "480p30", "360p30", "160p30", "worst"}

// func contains(qualitiesMap map[string]struct{}, quality string) bool {
// 	for  _, q := range(qualitiesMap) {
// 		if strings.HasPrefix()
// 	}
// }

func PrintQualities() string {
	return fmt.Sprintf("Supported qualities: [%s] ", strings.Join(supportedQualities, " "))
}

// GetQuality returns the appropriate quality from the list
func FindQuality(qualities []string, quality string) string {
	if len(qualities) == 0 {
		return quality
	}
	if quality == "best" {
		return qualities[0]
	}
	if quality == "worst" {
		return qualities[len(qualities)-1]
	}
	for i, q := range qualities {
		if strings.HasPrefix(q, quality) || strings.HasPrefix(quality, q) {
			return qualities[i]
		}
	}
	return quality
}

func IsValidQuality(q string) bool {
	for _, quality := range supportedQualities {
		if strings.HasPrefix(q, quality) || strings.HasPrefix(quality, q) {
			return true
		}
	}
	return false
}

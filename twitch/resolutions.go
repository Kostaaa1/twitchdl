package twitch

import (
	"strings"
)

type Resolution struct {
	res string
	fps int
}

var (
	resolutionKeys = []string{"chunked", "1080p60", "720p60", "480p30", "360p30", "160p30"}
	resolutions    = map[string]Resolution{
		"chunked": {res: "1920x1080", fps: 60},
		"1080p60": {res: "1920x1080", fps: 60},
		"720p60":  {res: "1280x720", fps: 60},
		"480p30":  {res: "854x480", fps: 30},
		"360p30":  {res: "640x360", fps: 30},
		"160p30":  {res: "284x160", fps: 30},
	}
)

func getResolution(quality string) string {
	if quality == "best" {
		return resolutionKeys[0]
	}
	if quality == "worst" {
		return resolutionKeys[len(resolutionKeys)-1]
	}
	for i, q := range resolutionKeys {
		if strings.HasPrefix(q, quality) || strings.HasPrefix(quality, q) {
			return resolutionKeys[i]
		}
	}
	return "1080p60"
}

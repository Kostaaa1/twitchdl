package m3u8

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type List struct {
	Bandwidth  string `json:"BANDWIDTH"`
	Resolution string `json:"RESOLUTION"`
	Video      string `json:"VIDEO"`
	FrameRate  string `json:"FRAME-RATE"`
	URL        string
}

type MasterPlaylist struct {
	UsherURL string
	Lists    []List
}

func extractSegmentInfo(mediaParts []string, seg *List) {
	structType := reflect.TypeOf(*seg)
	structValue := reflect.ValueOf(seg).Elem()
	for _, part := range mediaParts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		value := kv[1]
		value, err := strconv.Unquote(value)
		if err != nil {
			value = kv[1]
		}
		for structId := 0; structId < structType.NumField()-1; structId++ {
			field := structType.Field(structId)
			tag := field.Tag.Get("json")
			if key == tag {
				structFied := structValue.FieldByName(field.Name)
				if structFied.IsValid() && structFied.CanSet() {
					structFied.SetString(value)
				}
				break
			}
		}
	}
}

func Parse(playlist string) MasterPlaylist {
	var master MasterPlaylist
	lines := strings.Split(playlist, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			var segment List
			segment.URL = lines[i+1]
			mediaParts := strings.Split(strings.Split(line, ":")[1], ",")
			extractSegmentInfo(mediaParts, &segment)
			master.Lists = append(master.Lists, segment)
		}
	}
	return master
}

func (playlist *MasterPlaylist) GetMediaPlaylist(quality string) (List, error) {
	segments := playlist.Lists
	for i := 0; i < len(segments); i++ {
		seg := segments[i]
		if quality == "best" && seg.Video == "chunked" {
			return seg, nil
		}
		if seg.Video == quality {
			return seg, nil
		}
	}
	return List{}, fmt.Errorf("could not find the provided quality for a livestream")

}

func (playlist *MasterPlaylist) GetQualities() []string {
	segments := playlist.Lists
	var qualities []string
	for i := 0; i < len(segments); i++ {
		seg := segments[i]
		qualities = append(qualities, seg.Video)
		// if quality == "best" && seg.Video == "chunked" {
		// 	return seg, nil
		// }
		// if seg.Video == quality {
		// 	return seg, nil
		// }
	}
	return qualities
}

func (playlist *MasterPlaylist) GetJSONSegments() []string {
	var segments []string
	for _, seg := range playlist.Lists {
		b, err := json.MarshalIndent(seg, "", " ")
		if err != nil {
			break
		}
		segments = append(segments, string(b))
	}
	return segments
}

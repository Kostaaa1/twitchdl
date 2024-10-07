package m3u8

import (
	"encoding/json"
	"fmt"
	"strings"
)

// func extractSegmentInfo(mediaParts []string, seg *VariantPlaylist) {
// 	fmt.Println(mediaParts, seg)
// 	structType := reflect.TypeOf(*seg)
// 	structValue := reflect.ValueOf(seg).Elem()
// 	for _, part := range mediaParts {
// 		kv := strings.Split(part, "=")
// 		if len(kv) != 2 {
// 			continue
// 		}
// 		key := kv[0]
// 		value := kv[1]
// 		value, err := strconv.Unquote(value)
// 		if err != nil {
// 			value = kv[1]
// 		}
// 		for structId := 0; structId < structType.NumField()-1; structId++ {
// 			field := structType.Field(structId)
// 			tag := field.Tag.Get("json")
// 			if key == tag {
// 				structFied := structValue.FieldByName(field.Name)
// 				if structFied.IsValid() && structFied.CanSet() {
// 					structFied.SetString(value)
// 				}
// 				break
// 			}
// 		}
// 	}
// }

func New(fetchedPlaylist []byte) *MasterPlaylist {
	master := &MasterPlaylist{
		Serialized: string(fetchedPlaylist),
	}
	master.Parse()
	return master
}

func (m *MasterPlaylist) Parse() {
	lines := strings.Split(m.Serialized, "\n")

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			vl := ParseVariantPlaylist(line, lines[i+1])
			m.Lists = append(m.Lists, vl)

			i += 2
			if i >= len(lines) {
				break
			}
		}
	}
}

func (playlist *MasterPlaylist) GetVariantPlaylistByQuality(quality string) (VariantPlaylist, error) {
	mediaLists := playlist.Lists
	for i := 0; i < len(mediaLists); i++ {
		seg := mediaLists[i]
		if quality == "best" && seg.Video == "chunked" {
			return seg, nil
		}
		if strings.HasPrefix(seg.Video, quality) {
			return seg, nil
		}
	}

	return VariantPlaylist{}, fmt.Errorf("could not find the playlist by provided quality: %s", quality)
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

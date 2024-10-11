package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Kostaaa1/twitchdl/twitch"
)

func main() {
	// url := "https://www.twitch.tv/videos/2272036864"
	vodId := "2272036864"

	gqlPayload := `{
	    "query": "query { video(id: \"%s\") { broadcastType, createdAt, seekPreviewsURL, owner { login } } }"
	}`
	body := strings.NewReader(fmt.Sprintf(gqlPayload, vodId))

	tw := twitch.New()
	var p interface{}
	if err := tw.SendGqlLoadAndDecode(body, &p); err != nil {
		fmt.Println("Error sending GraphQL request:", err)
		return
	}

	b, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		fmt.Println("Error marshaling response:", err)
		return
	}
	fmt.Println(string(b))
}

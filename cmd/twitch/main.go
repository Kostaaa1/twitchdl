package main

import (
	"github.com/Kostaaa1/twitchdl/internal/config"
	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/view/chat"
)

func main() {
	jsonCfg, err := config.Get()
	if err != nil {
		panic(err)
	}

	tw := twitch.New()
	chat.Open(tw, jsonCfg)
	return
}

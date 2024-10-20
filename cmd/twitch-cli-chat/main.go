package main

import (
	"github.com/Kostaaa1/twitchdl/cmd/twitch-cli-chat/view/chat"
	"github.com/Kostaaa1/twitchdl/internal/config"
	"github.com/Kostaaa1/twitchdl/pkg/twitch"
)

func main() {
	jsonCfg, err := config.Get()
	if err != nil {
		panic(err)
	}
	tw := twitch.New()
	chat.Open(tw, jsonCfg)
}

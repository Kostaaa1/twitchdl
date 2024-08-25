package main

import (
	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/utils"
	"github.com/Kostaaa1/twitchdl/view/root"
)

func main() {
	cfg, err := utils.GetConfig()
	if err != nil {
		panic(err)
	}
	twitch := twitch.New()
	root.Open(twitch, cfg)
}

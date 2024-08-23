package main

import (
	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/utils"
	"github.com/Kostaaa1/twitchdl/view/prompts"
)

func main() {
	cfg, err := utils.GetConfig()
	if err != nil {
		panic(err)
	}
	twitch := twitch.New()

	prompts.Open(twitch, cfg)
}

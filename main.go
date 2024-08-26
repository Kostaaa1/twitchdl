package main

import (
	"github.com/Kostaaa1/twitchdl/internal/config"
	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/view/root"
)

func main() {
	cfg, err := config.Get()
	if err != nil {
		panic(err)
	}
	twitch := twitch.New()
	root.Open(twitch, cfg)
}

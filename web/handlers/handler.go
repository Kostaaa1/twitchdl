package handlers

import "github.com/Kostaaa1/twitchdl/pkg/twitch"

type Handler struct {
	twitch *twitch.Client
}

func New() *Handler {
	return &Handler{
		twitch: twitch.New(),
	}
}

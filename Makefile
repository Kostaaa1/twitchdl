.PHONY: run twitch-chat twitchdl web web-generate

FLAGS ?= ""

twitch-chat:
	go run ./cmd/twitch-cli-chat/main.go

twitchdl:
	go run ./cmd/twitch-cli-downloader/main.go $(FLAGS)

web:
	go run ./web/main.go

web-generate:
	templ generate && go run ./web/main.go

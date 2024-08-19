package types

type Chat struct {
	IsActive bool
	Channel  string
	Messages []string
	Room     Room
}

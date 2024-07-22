package main

import "github.com/Kostaaa1/twitchdl/chat"

var (
	name        = "slorpglorpski"
	accessToken = "x1ug4nduxyhopsdc1zrwbi1c3f5m0f"
	clientID    = "z4qytet5kietgqy0q7nxrgr8sverf1"
	secret      = "dzsedgplhczx5n0k25oj04339q0wei"
)

func main() {
	chat.Start()
	// msgChan := make(chan chat.UserMessage, 100)
	// ws, err := chat.CreateWSClient()
	// if err != nil {
	// 	panic(err)
	// }
	// go func() {
	// 	ws.Connect(accessToken, name, "hasanabi", msgChan)
	// }()
	// for msg := range msgChan {
	// 	fmt.Println(msg)
	// }

}

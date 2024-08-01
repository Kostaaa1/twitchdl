package main

import (
	"github.com/Kostaaa1/twitchdl/chat"
)

// var (
// 	name        = "slorpglorpski"
// 	accessToken = "x1ug4nduxyhopsdc1zrwbi1c3f5m0f"
// 	clientID    = "z4qytet5kietgqy0q7nxrgr8sverf1"
// 	secret      = "dzsedgplhczx5n0k25oj04339q0wei"
// )

func main() {
	chat.Start()

	// channels := []string{"nmplol", "kaellyn"}
	// at := "x1ug4nduxyhopsdc1zrwbi1c3f5m0f"
	// user := "slorpglorpski"
	// msgChan := make(chan interface{}, 100)
	// ws, err := chat.CreateWSClient()
	// if err != nil {
	// 	panic(err)
	// }
	// // Schedule an action to run after 10 seconds
	// time.AfterFunc(5*time.Second, func() {
	// 	ws.LeaveChannel("nmplol")
	// 	// fmt.Println("110 seconds have passed, performing the action10 seconds have passed, performing the action10 seconds have passed, performing the action10 seconds have passed, performing the action0 seconds have passed, performing the action...")
	// })
	// ws.Connect(at, user, msgChan, channels)
}

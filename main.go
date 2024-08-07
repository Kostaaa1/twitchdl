package main

import "github.com/Kostaaa1/twitchdl/view/chat"

func main() {
	chat.Open()
	///////////////////////////////////////////////////////////////
	// channels := []string{"emiru"}
	// at := "x1ug4nduxyhopsdc1zrwbi1c3f5m0f"
	// user := "slorpglorpski"
	// msgChan := make(chan interface{}, 100)
	// ws, err := chat.CreateWSClient()
	// if err != nil {
	// 	panic(err)
	// }
	// go ws.Connect(at, user, msgChan, channels)
	// for {
	// 	select {
	// 	case msg := <-msgChan:
	// 		fmt.Println(msg)
	// 	}
	// }
	///////////////////////////////////////////////////////////////
	// c := twitch.New(http.DefaultClient)
	// if err := c.OpenStreamInMediaPlayer("nmplol"); err != nil {
	// 	panic(err)
	// }
}

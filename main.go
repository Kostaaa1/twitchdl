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

	// msgChan := make(chan interface{}, 100)
	// ws, err := chat.CreateWSClient()
	// if err != nil {
	// 	panic(err)
	// }
	// ws.Connect(accessToken, name, "nmplol", msgChan)
	// for {
	// 	select {
	// 	case m := <-msgChan:
	// 		fmt.Println(m)
	// 	}
	// }
	// m := "@emote-only=0;followers-only=0;r9k=0;room-id=21841789;slow=0;subs-only=0 :tmi.twitch.tv ROOMSTATE #nmplol"
	// pattern := `\b(PRIVMSG|ROOMSTATE|USERNOTICE|USERSTATE|NOTICE|GLOBALUSERSTATE|CLEARMSG|CLEARCHAT)\b`
	// re := regexp.MustCompile(pattern)
	// k := re.FindStringSubmatch(m)
	// fmt.Println(k)

}

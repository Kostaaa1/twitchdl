package chat

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type UserMessage struct {
	Message        string
	Badges         []string
	Color          string
	DisplayName    string
	IsFirstMessage bool
	IsMod          bool
	IsSubscriber   bool
	ID             string
	Timestamp      string
	Type           string
}

type WebSocketClient struct {
	Conn *websocket.Conn
}

func CreateWSClient() (*WebSocketClient, error) {
	socketURL := "ws://irc-ws.chat.twitch.tv:80"
	conn, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
	if err != nil {
		return nil, err
	}
	return &WebSocketClient{Conn: conn}, nil
}

func (c *WebSocketClient) SendMessage(msg []byte) error {
	return c.Conn.WriteMessage(websocket.TextMessage, msg)
}

func parseZeroAndOne(v string) bool {
	n, _ := strconv.ParseInt(v, 10, 64)
	return n == 1
}

func parseTimestamp(v string) string {
	timestamp, _ := strconv.ParseInt(v, 10, 64)
	seconds := timestamp / 1000
	t := time.Unix(seconds, 0)
	formatted := t.Format("03:04 PM")
	return formatted
}

func ParseRawMessage(msg string) UserMessage {
	parts := strings.SplitN(msg, " :", 2)
	message := UserMessage{
		Message: strings.TrimSpace(strings.Split(parts[1], " :")[1]),
	}
	metadata := parts[0]
	kvPairs := strings.Split(metadata, ";")
	// fmt.Println(kvPairs)

	for _, pair := range kvPairs {
		kv := strings.Split(pair, "=")
		if len(kv) > 1 {
			key := kv[0]
			value := kv[1]

			switch key {
			case "color":
				message.Color = value
			case "display-name":
				message.DisplayName = value
			case "first-msg":
				message.IsFirstMessage = parseZeroAndOne(value)
			case "mod":
				message.IsMod = parseZeroAndOne(value)
			case "subscriber":
				message.IsSubscriber = parseZeroAndOne(value)
			case "user-id":
				message.ID = value
			case "user-type":
				message.Type = value
			case "tmi-sent-ts":
				message.Timestamp = parseTimestamp(value)
			}
		}
	}
	// b, err := json.MarshalIndent(message, "", " ")
	// if err != nil {
	// 	fmt.Println("failed to marshal indent: ", message)
	// }
	// fmt.Println(string(b))
	return message
}

func sendToChannel(msgChan chan UserMessage, msg UserMessage) {
	select {
	case msgChan <- msg:
		<-msgChan
		msgChan <- msg
	}
}

func (c *WebSocketClient) Connect(accessToken, username, channel string, msgChan chan UserMessage) {
	c.SendMessage([]byte("CAP REQ :twitch.tv/membership twitch.tv/tags twitch.tv/commands"))
	pass := fmt.Sprintf("PASS oauth:%s", accessToken)
	c.SendMessage([]byte(pass))
	nick := fmt.Sprintf("NICK %s", username)
	c.SendMessage([]byte(nick))
	join := fmt.Sprintf("JOIN #%s", channel)
	c.SendMessage([]byte(join))

	pattern := `\b(CLEARCHAT|CLEARMSG|GLOBALUSERSTATE|NOTICE|PRIVMSG|ROOMSTATE|USERNOTICE|USERSTATE)\b`
	re, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatalf("Error compiling regex: %v", err)
	}

	for {
		msgType, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			return
		}
		if msgType == websocket.TextMessage {
			rawIRCMessage := strings.TrimSpace(string(msg))
			tags := re.FindStringSubmatch(rawIRCMessage)
			if len(tags) > 1 {
				tag := tags[1]
				switch tag {
				case "PRIVMSG":
					parsedMsg := ParseRawMessage(rawIRCMessage)
					msgChan <- parsedMsg
				}
			}
		}
	}
}

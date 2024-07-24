package chat

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/types"
	"github.com/gorilla/websocket"
)

type WebSocketClient struct {
	Conn        *websocket.Conn
	CurrentUser types.UserIRC
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

func (c *WebSocketClient) FormatIRCMsgAndSend(tag, channel, msg string) error {
	formatted := fmt.Sprintf("%s #%s :%s", tag, channel, msg)
	return c.SendMessage([]byte(formatted))
}

func GetCurrentTimeFormatted() string {
	now := time.Now()
	timestamp := now.UnixNano() / int64(time.Millisecond)
	formattedTime := parseTimestamp(fmt.Sprintf("%d", timestamp))
	return formattedTime
}

func parseZeroAndOne(v string) bool {
	n, _ := strconv.ParseInt(v, 10, 64)
	return n == 1
}

func parseTimestamp(v string) string {
	timestamp, _ := strconv.ParseInt(v, 10, 64)
	seconds := timestamp / 1000
	t := time.Unix(seconds, 0)
	formatted := t.Format("03:04")
	return formatted
}

func parseROOMSTATE(rawMsg string) types.RoomState {
	var roomState types.RoomState
	var parts []string

	metadata := strings.Split(rawMsg, "@")
	if len(metadata) < 3 {
		return roomState
	}

	userMD := metadata[1]
	roomMD := metadata[2]
	parts = append(parts, strings.Split(userMD, ";")...)
	parts = append(parts, strings.Split(roomMD, ";")...)

	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) > 1 {
			key := kv[0]
			value := kv[1]
			switch key {
			case "display-name":
				roomState.DisplayName = value
			case "color":
				roomState.Color = value
			case "user-type":
				roomState.UserType = strings.Split(value, " :")[0]
			case "mod":
				roomState.IsMod = parseZeroAndOne(value)
			case "subscriber":
				roomState.IsSubscriber = parseZeroAndOne(value)
			case "emote-only":
				roomState.IsEmoteOnly = parseZeroAndOne(value)
			case "followers-only":
				roomState.IsFollowersOnly = parseZeroAndOne(value)
			case "subs-only":
				roomState.IsSubsOnly = parseZeroAndOne(value)
			case "room-id":
				roomState.RoomID = value
			}
		}
	}
	return roomState
}

func parsePRIVMSG(msg string) types.UserIRC {
	parts := strings.SplitN(msg, " :", 2)
	message := types.UserIRC{
		Message: strings.TrimSpace(strings.Split(parts[1], " :")[1]),
	}
	metadata := parts[0]
	kvPairs := strings.Split(metadata, ";")
	for _, pair := range kvPairs {
		kv := strings.Split(pair, "=")

		if len(kv) > 1 {
			key := kv[0]
			value := kv[1]
			switch key {
			// case "badges":
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
	//////////////////////////////
	// emojiRx := regexp.MustCompile(`[^\p{L}\p{N}\p{Zs}:/?&=.-]+`)
	emojiRx := regexp.MustCompile(`[^\p{L}\p{N}\p{Zs}:/?&=.-@]+`)
	message.Message = emojiRx.ReplaceAllString(message.Message, "")
	return message
}

func (c *WebSocketClient) Connect(accessToken, username, channel string, msgChan chan interface{}) {
	c.SendMessage([]byte("CAP REQ :twitch.tv/membership twitch.tv/tags twitch.tv/commands"))
	pass := fmt.Sprintf("PASS oauth:%s", accessToken)
	c.SendMessage([]byte(pass))
	nick := fmt.Sprintf("NICK %s", username)
	c.SendMessage([]byte(nick))
	join := fmt.Sprintf("JOIN #%s", channel)
	c.SendMessage([]byte(join))

	pattern := `\b(PRIVMSG|ROOMSTATE|USERNOTICE|USERSTATE|NOTICE|GLOBALUSERSTATE|CLEARMSG|CLEARCHAT)\b`
	re := regexp.MustCompile(pattern)

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
				case "USERSTATE":
					m1 := parseROOMSTATE(rawIRCMessage)
					msgChan <- m1
				case "PRIVMSG":
					parsed := parsePRIVMSG(rawIRCMessage)
					msgChan <- parsed
				}
			}
		}
	}
}

package chat

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/types"
	"github.com/Kostaaa1/twitchdl/utils"
	command "github.com/Kostaaa1/twitchdl/view/commands"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

type NewChannelMessage struct {
	Data interface{}
}

type Chat struct {
	IsActive bool
	channel  string
	messages []string
	room     types.Room
}

type Model struct {
	ws           *WebSocketClient
	viewport     viewport.Model
	labelBox     BoxWithLabel
	textinput    textinput.Model
	width        int
	height       int
	msgChan      chan interface{}
	chats        []Chat
	showCommands bool
	err          error
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

type InputCommand struct {
	cmd         string
	description string
}

var (
	commands = []InputCommand{
		{cmd: "/add", description: "Open new chat"},
		{cmd: "/close", description: "Closes the current chat"},
		{cmd: "/open", description: "Opens stream in your default media player"},
	}
)

func Open() {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	var data types.Config
	viper.Unmarshal(&data)

	if _, err := tea.
		NewProgram(initChatModel(data.Creds.AccessToken, data.DisplayName, data.ActiveChannels), tea.WithAltScreen()).
		Run(); err != nil {
		log.Fatal(err)
	}
}

func createNewChat(channel string, isActive bool) Chat {
	return Chat{
		IsActive: isActive,
		messages: []string{
			lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("Welcome to %s channel", channel)),
		},
		room:    types.Room{},
		channel: channel,
	}
}

func initChatModel(accessToken, username string, channels []string) tea.Model {
	vp := viewport.New(0, 0)
	vp.SetContent("")
	t := textinput.New()
	t.CharLimit = 500
	t.Placeholder = "Send a message"
	t.Prompt = "▶ "
	t.Focus()

	msgChan := make(chan interface{})
	ws, err := CreateWSClient()
	if err != nil {
		panic(err)
	}
	go ws.Connect(accessToken, username, msgChan, channels)

	chats := []Chat{}
	for i, channel := range channels {
		chats = append(chats, createNewChat(channel, i == 0))
	}
	return Model{
		ws:        ws,
		viewport:  vp,
		width:     0,
		height:    0,
		msgChan:   msgChan,
		labelBox:  NewBoxWithLabel("#8839ef"),
		textinput: t,
		chats:     chats,
	}
}

func (m Model) Init() tea.Cmd {
	return m.waitForMsg()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
	)
	m.textinput, tiCmd = m.textinput.Update(msg)
	m.showCommands = strings.HasPrefix(m.textinput.Value(), "/")

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width - 2
		h := msg.Height - 7
		m.labelBox.SetWidth(w)
		m.viewport.Width = w
		m.viewport.Height = h
		m.width = w
		m.height = h
		m.viewport.Style = lipgloss.
			NewStyle().
			Width(m.viewport.Width).
			Height(m.viewport.Height)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			m.sendMessage()
		case tea.KeyTab:
			m.nextTab()
		case tea.KeyShiftTab:
			m.prevTab()
		case tea.KeyCtrlW:
			if len(m.chats) > 1 {
				m.removeActiveChat()
			}
		case tea.KeyCtrlO:
			go func() {
				chat := m.getActiveChat()
				c := twitch.New(http.DefaultClient)
				if err := c.OpenStreamInMediaPlayer(chat.channel); err != nil {
					m.msgChan <- errMsg{err: err}

				}
			}()
		case tea.KeyShiftRight:
			m.width = (m.width / 3) * 2
			command.Open(m.width/3, m.height)
		}

	case errMsg:
		m.err = msg.err
		return m, m.waitForMsg()

	case NewChannelMessage:
		switch chanMsg := msg.Data.(type) {
		case types.Room:
			m.addRoomToChat(chanMsg)
		case types.ChatMessage:
			chat := m.getChat(chanMsg.Metadata.RoomID)
			if chat != nil {
				m.appendMessage(chat, FormatChatMessage(chanMsg, m.width))
			}
		case types.SubNotice:
			chat := m.getChat(chanMsg.Metadata.RoomID)
			if chat != nil {
				m.appendMessage(chat, FormatSubMessage(chanMsg, m.width))
			}
		case types.Notice:
			chat := m.getChat(chanMsg.DisplayName)
			if chat != nil {
				m.appendMessage(chat, chanMsg.SystemMsg)
			}
		}
		return m, m.waitForMsg()
	}
	return m, tea.Batch(tiCmd)
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.labelBox.
		SetWidth(m.viewport.Width).
		RenderBoxWithTabs(m.chats, m.viewport.View()))
	b.WriteString(fmt.Sprintf("%s ", m.renderRoom()))
	b.WriteString(m.textinput.View())
	b.WriteString(m.renderError())
	return b.String()
}

func (m *Model) createNewMessage(chat *Chat) types.ChatMessage {
	newMessage := types.ChatMessage{
		Message: m.textinput.Value(),
		Metadata: types.ChatMessageMetadata{
			Metadata: types.Metadata{
				Color:        chat.room.Metadata.Color,
				DisplayName:  chat.room.Metadata.DisplayName,
				IsMod:        chat.room.Metadata.IsMod,
				IsSubscriber: chat.room.Metadata.IsSubscriber,
				UserType:     chat.room.Metadata.UserType,
			},
			RoomID:    chat.room.RoomID,
			Timestamp: utils.GetCurrentTimeFormatted(),
		},
	}
	return newMessage
}

func (m Model) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		newMsg := <-m.msgChan
		switch newMsg.(type) {
		case errMsg:
			time.AfterFunc(time.Second*2, func() {
				m.msgChan <- errMsg{err: nil}
			})
			return newMsg
		default:
			return NewChannelMessage{Data: newMsg}
		}
	}
}

func (m *Model) renderError() string {
	var b strings.Builder
	if m.err != nil {
		b.WriteString(lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("\n\n[Error] - %s", m.err)))
	} else {
		b.WriteString("")
	}
	return b.String()
}

func (m *Model) sendMessage() {
	if m.textinput.Value() == "" {
		return
	}
	input := m.textinput.Value()
	if !strings.HasPrefix(input, "/") {
		chat := m.getActiveChat()
		newMessage := m.createNewMessage(chat)
		m.ws.FormatIRCMsgAndSend("PRIVMSG", chat.channel, input)
		chat.messages = append(chat.messages, FormatChatMessage(newMessage, m.width))
		m.updateChatViewport(chat)
	} else {
		m.handleInputCommand(input)
	}
	m.textinput.Reset()
}

func (m *Model) handleInputCommand(cmd string) {
	parts := strings.Split(cmd, " ")
	if len(parts) > 2 {
		return
	}
	switch parts[0] {
	case "/add":
		newChat := createNewChat(parts[1], false)
		m.addChat(newChat)
	default:
		m.msgChan <- errMsg{err: fmt.Errorf("invalid command: %s", cmd)}
	}
}

func (m *Model) addChat(newChat Chat) {
	m.chats = append(m.chats, newChat)
	m.ws.ConnectToChannel(newChat.channel)
	newChannels := []string{}
	for _, c := range m.chats {
		newChannels = append(newChannels, c.channel)
	}
	viper.Set("activeChannels", newChannels)
	viper.WriteConfig()
}

func (m *Model) addRoomToChat(chanMsg types.Room) {
	for i := range m.chats {
		c := &(m.chats)[i]
		if c.channel == chanMsg.Metadata.Channel {
			c.room = chanMsg
			break
		}
	}
}

func (m *Model) removeActiveChat() {
	var activeChan string
	chats := []Chat{}
	newActiveId := 0

	for i, c := range m.chats {
		if !c.IsActive {
			chats = append(chats, c)
		} else {
			activeChan = c.channel
			m.ws.LeaveChannel(c.channel)
			newActiveId = i
			if i == len(m.chats)-1 {
				newActiveId--
			}
		}
	}
	chats[newActiveId].IsActive = true
	newActiveC := chats[newActiveId]
	m.updateChatViewport(&newActiveC)

	// remove from config...
	activeChans := viper.GetStringSlice("activeChannels")
	newActiveChannels := []string{}
	for _, ch := range activeChans {
		if ch != activeChan {
			newActiveChannels = append(newActiveChannels, ch)
		}
	}
	viper.Set("activeChannels", newActiveChannels)
	viper.WriteConfig()

	m.chats = chats
}

func (m *Model) appendMessage(chat *Chat, message string) {
	if len(chat.messages) > 100 {
		chat.messages = chat.messages[1:]
	}
	chat.messages = append(chat.messages, message)
	if chat.IsActive {
		m.updateChatViewport(chat)
	}
}

func (m *Model) updateChatViewport(chat *Chat) {
	m.viewport.SetContent(strings.Join(chat.messages, "\n"))
	m.viewport.GotoBottom()
}

func (m *Model) nextTab() {
	var activeIndex int
	for i, chat := range m.chats {
		if chat.IsActive {
			activeIndex = i
			break
		}
	}
	(m.chats)[activeIndex].IsActive = false
	nextIndex := (activeIndex + 1) % len(m.chats)
	(m.chats)[nextIndex].IsActive = true
	m.updateChatViewport(&(m.chats)[nextIndex])
}

func (m *Model) prevTab() {
	var activeIndex int
	for i, c := range m.chats {
		if c.IsActive {
			activeIndex = i
			break
		}
	}
	(m.chats)[activeIndex].IsActive = false
	prevIndex := (activeIndex - 1 + len(m.chats)) % len(m.chats)
	(m.chats)[prevIndex].IsActive = true
	m.updateChatViewport(&(m.chats)[prevIndex])
}

func (m Model) getActiveChat() *Chat {
	for i := range m.chats {
		if (m.chats)[i].IsActive {
			return &(m.chats[i])
		}
	}
	return nil
}

func (m Model) getChat(roomID string) *Chat {
	for i := range m.chats {
		if (m.chats)[i].room.RoomID == roomID || (m.chats)[i].channel == roomID {
			return &(m.chats[i])
		}
	}
	return nil
}

func (m Model) renderRoom() string {
	chat := m.getActiveChat()
	style := lipgloss.NewStyle().Faint(true)
	switch {
	case chat.room.IsEmoteOnly:
		return style.Render("[Emote-Only Chat]")
	case chat.room.IsFollowersOnly:
		return style.Render("[Followers-Only Chat]")
	case chat.room.IsSubsOnly:
		return style.Render("[Subscriber-Only Chat]")
	default:
		return ""
	}
}

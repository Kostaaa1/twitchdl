package chat

import (
	"fmt"
	"log"
	"strings"

	"github.com/Kostaaa1/twitchdl/types"
	"github.com/Kostaaa1/twitchdl/utils"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NewChannelMessage struct {
	Data interface{}
}

type Chat struct {
	// ID          int
	IsActive    bool
	messages    []string
	channelName string
	roomState   types.Room
}

type Model struct {
	ws        *WebSocketClient
	viewport  viewport.Model
	labelBox  BoxWithLabel
	textinput textinput.Model
	width     int
	height    int
	msgChan   chan interface{}
	chats     *[]Chat
}

func Start() {
	channels := []string{"nmplol", "zackrawrr", "kaellyn"}
	at := "x1ug4nduxyhopsdc1zrwbi1c3f5m0f"
	user := "slorpglorpski"

	if _, err := tea.NewProgram(initModel(at, user, channels), tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}

func initModel(accessToken, username string, channels []string) tea.Model {
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
		chats = append(chats, Chat{
			// ID:          i,
			IsActive:    i == 0,
			messages:    []string{},
			roomState:   types.Room{},
			channelName: channel,
		})
	}
	return Model{
		ws:        ws,
		viewport:  vp,
		width:     0,
		height:    0,
		msgChan:   msgChan,
		labelBox:  NewBoxWithLabel("#8839ef"),
		textinput: t,
		chats:     &chats,
	}
}

func (m Model) getActiveChat() *Chat {
	for i := range *m.chats {
		if (*m.chats)[i].IsActive {
			c := &(*m.chats)[i]
			return c
		}
	}
	return nil
}

func (m Model) getChat(roomID string) *Chat {
	for i := range *m.chats {
		if (*m.chats)[i].roomState.RoomID == roomID {
			c := &(*m.chats)[i]
			return c
		}
	}
	return nil
}

func (m Model) Init() tea.Cmd {
	return m.waitForMsg()
}

func (m Model) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		newMsg := <-m.msgChan
		return NewChannelMessage{Data: newMsg}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)
	m.textinput, tiCmd = m.textinput.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width - 2
		h := msg.Height - 7
		m.labelBox.SetWidth(w)
		m.viewport.Width = w
		m.viewport.Height = h
		m.width = w
		m.height = h
		m.viewport.Style = lipgloss.NewStyle().
			Width(m.viewport.Width).
			Height(m.viewport.Height)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textinput.Value() == "" {
				return m, nil
			}
			m.sendMessage()
		case tea.KeyTab:
			m.nextTab()
		case tea.KeyShiftTab:
			m.prevTab()
		case tea.KeyCtrlX:
			m.removeActiveChat()
		case tea.KeyUp, tea.KeyDown:
			m.viewport, vpCmd = m.viewport.Update(msg)
		}

	case NewChannelMessage:
		switch chanMsg := msg.Data.(type) {
		case types.Room:
			m.addRoom(chanMsg)
		case types.ChatMessage:
			chat := m.getChat(chanMsg.Metadata.RoomID)
			m.addMessageToChat(chat, FormatChatMessage(chanMsg, m.width))
		case types.SubNotice:
			chat := m.getChat(chanMsg.Metadata.RoomID)
			m.addMessageToChat(chat, FormatSubMessage(chanMsg, m.width))
		}
		return m, m.waitForMsg()
	}
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m *Model) removeActiveChat() {
	if len(*m.chats) == 1 {
		return
	}

	chat := m.getActiveChat()
	m.ws.LeaveChannel(chat.channelName)
	chats := []Chat{}
	for i := range *m.chats {
		c := &(*m.chats)[i]
		if !c.IsActive {
			chats = append(chats, *c)
		} else {
			if i == 0 {
				nextC := &(*m.chats)[i+1]
				nextC.IsActive = true
				chats = append(chats, *nextC)
				m.updateViewport(nextC)
			} else {
				chats[i-1].IsActive = true
				m.updateViewport(&chats[i-1])
			}
		}
	}
	m.chats = &chats
}

func (m *Model) sendMessage() {
	chat := m.getActiveChat()
	newMessage := types.ChatMessage{
		Message: m.textinput.Value(),
		Metadata: types.ChatMessageMetadata{
			Metadata: types.Metadata{
				Color:        chat.roomState.Metadata.Color,
				DisplayName:  chat.roomState.Metadata.DisplayName,
				IsMod:        chat.roomState.Metadata.IsMod,
				IsSubscriber: chat.roomState.Metadata.IsSubscriber,
				UserType:     chat.roomState.Metadata.UserType,
			},
			RoomID:    chat.roomState.RoomID,
			Timestamp: utils.GetCurrentTimeFormatted(),
		},
	}
	m.ws.FormatIRCMsgAndSend("PRIVMSG", chat.channelName, m.textinput.Value())
	chat.messages = append(chat.messages, FormatChatMessage(newMessage, m.width))
	m.viewport.SetContent(strings.Join(chat.messages, "\n"))
	m.textinput.Reset()
	m.viewport.GotoBottom()
}

func (m *Model) addRoom(chanMsg types.Room) {
	for i := range *m.chats {
		c := &(*m.chats)[i]
		if c.roomState.RoomID == "" {
			c.roomState = chanMsg
			initMsg := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("Welcome to %s channel", c.channelName))
			c.messages = append(c.messages, initMsg)
			break
		}
	}
}

func (m *Model) addMessageToChat(chat *Chat, message string) {
	if len(chat.messages) > 100 {
		chat.messages = chat.messages[1:]
	}
	chat.messages = append(chat.messages, message)
	if chat.IsActive {
		m.updateViewport(chat)
	}
}

func (m *Model) updateViewport(chat *Chat) {
	m.viewport.SetContent(strings.Join(chat.messages, "\n"))
	m.viewport.GotoBottom()
}

func (m *Model) nextTab() {
	for i := range *m.chats {
		if i < len(*m.chats)-1 && (*m.chats)[i].IsActive {
			c := &(*m.chats)[i]
			c.IsActive = false
			next := &(*m.chats)[i+1]
			next.IsActive = true
			m.updateViewport(next)
			break
		}
	}
	// t := m.getActiveChat()
	// if t.ID < len(*m.chats)-1 {
	// 	t.IsActive = false
	// 	next := &(*m.chats)[t.ID+1]
	// 	next.IsActive = true
	// 	m.updateViewport(next)
	// }
}

func (m *Model) prevTab() {
	for i := range *m.chats {
		if i > 0 && (*m.chats)[i].IsActive {
			c := &(*m.chats)[i]
			c.IsActive = false
			prevChat := &(*m.chats)[i-1]
			prevChat.IsActive = true
			m.updateViewport(prevChat)
			break
		}
	}
}

func (m Model) renderRoomState() string {
	chat := m.getActiveChat()
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#888892"))

	switch {
	case chat.roomState.IsEmoteOnly:
		return style.Render("[Emote-Only Chat] ")
	case chat.roomState.IsFollowersOnly:
		return style.Render("[Followers-Only Chat] ")
	case chat.roomState.IsSubsOnly:
		return style.Render("[Subscriber-Only Chat] ")
	default:
		return " "
	}
}

func renderInfo() string {
	return lipgloss.NewStyle().Faint(true).Render("\n\n[Ctrl+Tab]-next [Shfit+Tab]-prev [Ctrl+x]-Close chat")
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.labelBox.
		SetWidth(m.viewport.Width).
		RenderBoxWithTabs(m.chats, m.viewport.View()))
	b.WriteString(m.renderRoomState())
	b.WriteString(m.textinput.View())
	b.WriteString(renderInfo())
	return b.String()
}
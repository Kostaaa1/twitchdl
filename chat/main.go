package chat

import (
	"fmt"
	"log"
	"strings"

	"github.com/Kostaaa1/twitchdl/types"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NewChannelMessage struct {
	Data interface{}
}

type Chat struct {
	ID        int
	IsActive  bool
	messages  []string
	channel   string
	roomState types.RoomState
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
	if _, err := tea.NewProgram(initModel(), tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}

func initModel() tea.Model {
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

	// channel := "zackrawrr"
	// channel2 := "piratesoftware"
	channels := []string{"zackrawrr", "hasanabi"}
	go ws.Connect("x1ug4nduxyhopsdc1zrwbi1c3f5m0f", "slorpglorpski", msgChan, channels)

	return Model{
		ws:        ws,
		viewport:  vp,
		width:     0,
		height:    0,
		msgChan:   msgChan,
		labelBox:  NewBoxWithLabel("#8839ef"),
		textinput: t,
		chats: &[]Chat{
			{
				ID:        0,
				IsActive:  true,
				roomState: types.RoomState{},
				messages:  []string{},
				channel:   channels[0],
			},
			{
				ID:        1,
				IsActive:  false,
				roomState: types.RoomState{},
				messages:  []string{},
				channel:   channels[1],
			},
		},
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
	// chat := &(*m.chats)[0]

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

			// newMessage := types.ChatMessage{
			// 	Message: m.textinput.Value(),
			// 	// Color:        m.roomState.Color,
			// 	// DisplayName:  m.roomState.DisplayName,
			// 	// IsMod:        m.roomState.IsMod,
			// 	// IsSubscriber: m.roomState.IsSubscriber,
			// 	// Timestamp:    utils.GetCurrentTimeFormatted(),
			// }
			// m.ws.FormatIRCMsgAndSend("PRIVMSG", chat.channel, m.textinput.Value())
			// chat.messages = append(chat.messages, FormatChatMessage(newMessage, m.width))
			// m.viewport.SetContent(strings.Join(chat.messages, "\n"))
			// m.textinput.Reset()
			// m.viewport.GotoBottom()

		case tea.KeyTab:
			t := m.getActiveChat()
			if t.ID < len(*m.chats)-1 {
				t.IsActive = false
				next := &(*m.chats)[t.ID+1]
				next.IsActive = true
				m.viewport.SetContent(strings.Join(next.messages, "\n"))
				m.viewport.GotoBottom()
				return m, m.waitForMsg()
			}

		case tea.KeyShiftTab:
			t := m.getActiveChat()
			if t.ID > 0 {
				t.IsActive = false
				prev := &(*m.chats)[t.ID-1]
				prev.IsActive = true
				m.viewport.SetContent(strings.Join(prev.messages, "\n"))
				m.viewport.GotoBottom()
				return m, m.waitForMsg()
			}
		case tea.KeyUp, tea.KeyDown:
			m.viewport, vpCmd = m.viewport.Update(msg)
		}

	case NewChannelMessage:
		switch chanMsg := msg.Data.(type) {
		case types.RoomState:
			for i := range *m.chats {
				c := &(*m.chats)[i]
				if c.roomState.RoomID == "" {
					c.roomState = chanMsg
					initMsg := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("Welcome to %s channel", c.channel))
					c.messages = append(c.messages, initMsg)
					break
				}
			}
			return m, m.waitForMsg()

		case types.ChatMessage:
			chat := m.getChat(chanMsg.Metadata.RoomID)
			if len(chat.messages) == 100 {
				chat.messages = chat.messages[1:]
			}
			chat.messages = append(chat.messages, FormatChatMessage(chanMsg, m.width))
			if chat.IsActive {
				m.viewport.SetContent(strings.Join(chat.messages, "\n"))
				m.viewport.GotoBottom()
			}
			return m, m.waitForMsg()

		case types.SubNotice:
			chat := m.getChat(chanMsg.Metadata.RoomID)
			if len(chat.messages) == 100 {
				chat.messages = chat.messages[1:]
			}
			chat.messages = append(chat.messages, FormatSubMessage(chanMsg, m.width))
			if chat.IsActive {
				m.viewport.SetContent(strings.Join(chat.messages, "\n"))
				m.viewport.GotoBottom()
			}
			return m, m.waitForMsg()
		}
	}
	return m, tea.Batch(tiCmd, vpCmd)
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

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.labelBox.
		SetWidth(m.viewport.Width).
		RenderBoxWithTabs(m.chats, m.viewport.View()))
	b.WriteString(m.renderRoomState())
	b.WriteString(m.textinput.View())
	return b.String()
}

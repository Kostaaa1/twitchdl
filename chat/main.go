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

type Model struct {
	ws        *WebSocketClient
	viewport  viewport.Model
	width     int
	height    int
	msgChan   chan interface{}
	textinput textinput.Model
	messages  []string
	labelBox  utils.BoxWithLabel
	channel   string
	roomState types.RoomState
}

func Start() {
	if _, err := tea.NewProgram(initModel(), tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}

func initModel() tea.Model {
	channel := "loud_coringa"
	vp := viewport.New(0, 0)
	vp.SetContent("")

	msgChan := make(chan interface{})
	ws, err := CreateWSClient()
	if err != nil {
		panic(err)
	}
	go ws.Connect("x1ug4nduxyhopsdc1zrwbi1c3f5m0f", "slorpglorpski", channel, msgChan)

	t := textinput.New()
	t.CharLimit = 500
	t.Placeholder = "Send a message"
	t.Prompt = "▶ "
	t.Focus()
	labelBox := utils.NewBoxWithLabel("#8839ef")

	return Model{
		ws:        ws,
		viewport:  vp,
		width:     0,
		height:    0,
		msgChan:   msgChan,
		textinput: t,
		roomState: types.RoomState{},
		labelBox:  labelBox,
		messages:  []string{},
		channel:   channel,
	}
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
			newMessage := types.ChatMessage{
				Message: m.textinput.Value(),
				// Color:        m.roomState.Color,
				// DisplayName:  m.roomState.DisplayName,
				// IsMod:        m.roomState.IsMod,
				// IsSubscriber: m.roomState.IsSubscriber,
				// Timestamp:    utils.GetCurrentTimeFormatted(),
			}
			m.ws.FormatIRCMsgAndSend("PRIVMSG", m.channel, m.textinput.Value())
			m.messages = append(m.messages, utils.FormatChatMessage(newMessage, m.width))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textinput.Reset()
			m.viewport.GotoBottom()
		case tea.KeyUp, tea.KeyDown:
			m.viewport, vpCmd = m.viewport.Update(msg)
		}

	case NewChannelMessage:
		switch chanMsg := msg.Data.(type) {
		case types.RoomState:
			m.roomState = chanMsg
			m.messages = append(m.messages, lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("Welcome to %s channel", m.channel)))
			return m, m.waitForMsg()
		case types.ChatMessage:
			if len(m.messages) == 100 {
				m.messages = m.messages[1:]
			}
			m.messages = append(m.messages, utils.FormatChatMessage(chanMsg, m.width))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			return m, m.waitForMsg()
		case types.SubNotice:
			m.messages = append(m.messages, utils.FormatSubMessage(chanMsg, m.width))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			return m, m.waitForMsg()
		}
	}
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) renderRoomState() string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#888892"))
	switch {
	case m.roomState.IsEmoteOnly:
		return style.Render("[Emote-Only Chat] ")
	case m.roomState.IsFollowersOnly:
		return style.Render("[Followers-Only Chat] ")
	case m.roomState.IsSubsOnly:
		return style.Render("[Subscriber-Only Chat] ")
	default:
		return " "
	}
}

func (m Model) View() string {
	var b strings.Builder
	b.WriteString(m.labelBox.
		SetWidth(m.viewport.Width).
		RenderBoxWithTabs(fmt.Sprintf(" %s ", m.channel), m.viewport.View()))
	b.WriteString(m.renderRoomState())
	b.WriteString(m.textinput.View())
	return b.String()
}

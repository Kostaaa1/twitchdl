package chat

import (
	"fmt"
	"log"
	"math/rand"
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

type ChatModel struct {
	ws        *WebSocketClient
	msgChan   chan interface{}
	roomState types.RoomState
	textinput textinput.Model
	viewport  viewport.Model
	width     int
	height    int
	messages  []string
}

var twitchChan = "zackrawrr"

func Start() {
	if _, err := tea.NewProgram(initModel(), tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}

func initModel() tea.Model {
	vp := viewport.New(0, 0)
	vp.SetContent("")

	msgChan := make(chan interface{})
	ws, err := CreateWSClient()
	if err != nil {
		panic(err)
	}
	go ws.Connect("x1ug4nduxyhopsdc1zrwbi1c3f5m0f", "slorpglorpski", twitchChan, msgChan)

	ta := textinput.New()
	ta.CharLimit = 500
	ta.Placeholder = "Send a message"
	ta.Prompt = "▶ "
	ta.Focus()

	return ChatModel{
		ws:        ws,
		roomState: types.RoomState{},
		textinput: ta,
		viewport:  vp,
		msgChan:   msgChan,
		width:     0,
		height:    0,
		messages:  []string{},
	}
}

func (m ChatModel) Init() tea.Cmd {
	return m.waitForMsg()
}

func (m ChatModel) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		return NewChannelMessage{Data: <-m.msgChan}
	}
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)
	m.textinput, tiCmd = m.textinput.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width - 2
		h := msg.Height - 7

		m.viewport.Width = w
		m.viewport.Height = h
		m.width = w
		m.height = h
		m.viewport.Style = lipgloss.NewStyle().
			Width(m.viewport.Width).
			Height(m.viewport.Height).
			Padding(0, 1, 0, 1).
			MarginBottom(1).
			BorderStyle(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("63"))

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			if m.textinput.Value() == "" {
				return m, nil
			}
			newMsg := types.UserIRC{
				DisplayName:    "Kosta",
				Badges:         []string{},
				Color:          "#FFF200",
				IsFirstMessage: false,
				IsMod:          false,
				IsSubscriber:   false,
				Type:           "",
				ID:             "93289321",
				Message:        m.textinput.Value(),
				Timestamp:      GetCurrentTimeFormatted(),
			}
			m.ws.FormatIRCMsgAndSend("PRIVMSG", twitchChan, m.textinput.Value())
			m.messages = append(m.messages, formatMsg(newMsg, m.width))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textinput.Reset()
			m.viewport.GotoBottom()

		case tea.KeyUp, tea.KeyDown:
			m.viewport, vpCmd = m.viewport.Update(msg)
		}

	case NewChannelMessage:
		// fmt.Println(msg.Data)
		switch chanMsg := msg.Data.(type) {
		case types.RoomState:
			m.roomState = chanMsg
			return m, m.waitForMsg()
		case types.UserIRC:
			if len(m.messages) == 100 {
				m.messages = m.messages[1:]
			}
			m.messages = append(m.messages, formatMsg(chanMsg, m.width-10))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
			return m, m.waitForMsg()
		}
	}
	return m, tea.Batch(tiCmd, vpCmd)
}

func formatMsg(msg types.UserIRC, width int) string {
	trimmedMsg := strings.TrimSpace(msg.Message)
	if msg.Color == "" {
		msg.Color = getRandHex()
	}
	return fmt.Sprintf("[%s] %s: %s", msg.Timestamp, lipgloss.NewStyle().Foreground(lipgloss.Color(msg.Color)).Render(msg.DisplayName), trimmedMsg)
}

func (m ChatModel) View() string {
	return fmt.Sprintf(
		"%s\n%s",
		m.viewport.View(),
		m.textinput.View(),
	)
}

func getRandHex() string {
	getHex := func(rgb int) string {
		hex := fmt.Sprintf("%x", rgb)
		if len(hex) == 1 {
			hex = "0" + hex
		}
		return hex
	}
	rgb := struct {
		Red   int
		Green int
		Blue  int
	}{
		Red:   rand.Intn(500),
		Green: rand.Intn(500),
		Blue:  rand.Intn(500),
	}
	hex := fmt.Sprintf("#%s%s%s", getHex(rgb.Red), getHex(rgb.Green), getHex(rgb.Blue))
	return hex
}

package user

import (
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	twitch              *twitch.Client
	viewport            viewport.Model
	textinput           textinput.Model
	width               int
	height              int
	msgChan             chan interface{}
	showCommands        bool
	commandsWindowWidth int
	err                 error
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

type NewChannelMessage struct {
	Data interface{}
}

func Open() {
	if _, err := tea.NewProgram(initChatModel(), tea.WithAltScreen()).Run(); err != nil {
		panic(err)
	}
}

func initChatModel() tea.Model {
	// cfg, err := utils.GetConfig()
	// if err != nil {
	// 	panic(err)
	// }

	vp := viewport.New(0, 0)
	vp.SetContent("")
	t := textinput.New()
	t.CharLimit = 500
	t.Placeholder = "Send a message"
	t.Prompt = " ▶ "
	t.Focus()

	msgChan := make(chan interface{})
	// ws, err := CreateWSClient()
	// if err != nil {
	// 	panic(err)
	// }
	// go ws.Connect(cfg.Creds.AccessToken, cfg.Creds.ClientID, msgChan, cfg.ActiveChannels)

	return Model{
		twitch:              twitch.New(),
		err:                 nil,
		width:               0,
		height:              0,
		msgChan:             msgChan,
		viewport:            vp,
		textinput:           t,
		showCommands:        false,
		commandsWindowWidth: 30,
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

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width - 2
		h := msg.Height - 8
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
		case tea.KeyTab:
			m.showCommands = !m.showCommands
			if m.showCommands {
				m.viewport.Width = m.width - m.commandsWindowWidth
			} else {
				m.viewport.Width = m.width
			}
		}

	case errMsg:
		m.err = msg.err
		return m, m.waitForMsg()

	case NewChannelMessage:
		return m, m.waitForMsg()
	}
	return m, tea.Batch(tiCmd)
}

func (m Model) View() string {
	var b strings.Builder
	return b.String()
}

func (m Model) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		newMsg := <-m.msgChan
		switch newMsg.(type) {
		case errMsg:
			time.AfterFunc(time.Second*3, func() {
				m.msgChan <- errMsg{err: nil}
			})
			return newMsg
		default:
			return NewChannelMessage{Data: newMsg}
		}
	}
}

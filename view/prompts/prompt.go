package prompts

import (
	"strings"

	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type NewChannelMessage struct {
	Data interface{}
}

type model struct {
	twitch    *twitch.Client
	viewport  viewport.Model
	textinput textinput.Model

	// ws                  *WebSocketClient
	// labelBox components.BoxWithLabel
	// width               int
	// height              int
	// msgChan             chan interface{}
	// chats               []types.Chat
	// showCommands        bool
	// commandsWindowWidth int
	// err                 error
	// mu                  *sync.Mutex

}

func Open() {
	if _, err := tea.NewProgram(initChatModel(), tea.WithAltScreen()).Run(); err != nil {
		panic(err)
	}
}

func initChatModel() tea.Model {
	vp := viewport.New(0, 0)
	vp.SetContent("")

	t := textinput.New()
	t.CharLimit = 500
	t.Placeholder = " url"
	t.Prompt = "Enter clip or video URL: "
	t.Cursor.Blink = true
	t.Focus()

	return model{
		twitch:    twitch.New(),
		viewport:  vp,
		textinput: t,
		// ws:                  ws,
		// chats:               chats,
		// err:                 nil,
		// width:               0,
		// height:              0,
		// msgChan:             msgChan,
		// labelBox:            components.NewBoxWithLabel("63"),
		// showCommands:        false,
		// commandsWindowWidth: 32,
	}
}

func (m model) Init() tea.Cmd {
	return tea.ShowCursor
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.viewport.Style = lipgloss.
			NewStyle().
			Width(m.viewport.Width).
			Height(m.viewport.Height)

		// if m.chats[0].IsActive {
		// 	m.updateChatViewport(&m.chats[0])
		// }

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	return m, tea.Batch(tiCmd)
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(m.textinput.View())
	return b.String()
}

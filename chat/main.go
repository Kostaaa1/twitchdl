package chat

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type UpdateMsg struct {
	Data UserMessage
}

type ChatModel struct {
	input    string
	msgChan  chan UserMessage
	textarea textarea.Model
	viewport viewport.Model
	width    int
	height   int
	messages []string
}

// type model struct {
// 	messages []string
// 	textarea textarea.Model
// 	viewport viewport.Model
// 	msgChan  chan UserMessage
// 	err      error
// }

func Start() {
	if _, err := tea.NewProgram(initModel(), tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}

func initModel() tea.Model {
	vp := viewport.New(0, 0)
	vp.SetContent("")
	msgChan := make(chan UserMessage, 20)

	ws, err := CreateWSClient()
	if err != nil {
		panic(err)
	}
	go ws.Connect("x1ug4nduxyhopsdc1zrwbi1c3f5m0f", "slorpglorpski", "hasanabi", msgChan)

	ta := textarea.New()
	ta.CharLimit = 500
	ta.Placeholder = "Send a message"
	ta.Prompt = "    ▶ "
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)
	ta.Focus()

	return ChatModel{
		input:    "",
		textarea: ta,
		viewport: vp,
		msgChan:  msgChan,
		width:    0,
		height:   0,
		messages: []string{},
	}

}

func (m ChatModel) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.waitForMsg())
}

func (m ChatModel) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		return UpdateMsg{Data: <-m.msgChan}
	}
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)
	m.textarea, tiCmd = m.textarea.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// v, h := docStyle.GetFrameSize()
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(1)

		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 7
		m.viewport, vpCmd = m.viewport.Update(msg)
		m.viewport.Style = lipgloss.NewStyle().Padding(0, 1).MarginBottom(1).MarginLeft(3).BorderStyle(lipgloss.ThickBorder()).BorderForeground(lipgloss.Color("61"))

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			// send message.
			if m.textarea.Value() == "" {
				return m, nil
			}

			newMsg := struct {
				Color lipgloss.Style
				Text  string
				Name  string
			}{
				Color: lipgloss.NewStyle().Foreground(lipgloss.Color("#FFF200")),
				Text:  m.textarea.Value(),
				Name:  "Kosta",
			}

			msgLine := fmt.Sprintf("%s: %s", newMsg.Color.Render(newMsg.Name), m.textarea.Value())
			m.messages = append(m.messages, msgLine)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()

		case tea.KeyUp, tea.KeyDown:
			m.viewport, vpCmd = m.viewport.Update(msg)
		}

	case UpdateMsg:
		if len(m.messages) == 100 {
			m.messages = m.messages[50:]
		}

		msgLine := fmt.Sprintf("%s: %s", lipgloss.NewStyle().Foreground(lipgloss.Color(msg.Data.Color)).Render(msg.Data.DisplayName), msg.Data.Message)
		m.messages = append(m.messages, msgLine)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

		return m, m.waitForMsg()
	}
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m ChatModel) View() string {
	// return m.viewport.View()
	return fmt.Sprintf(
		"%s\n%s",
		m.viewport.View(),
		m.textarea.View(),
	)
}

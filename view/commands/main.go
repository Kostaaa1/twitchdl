package command

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	viewport viewport.Model
	// textinput textinput.Model
	width  int
	height int
}

func Open(width, height int) {
	if _, err := tea.NewProgram(initModel(width, height), tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}

func initModel(width, height int) tea.Model {
	vp := viewport.New(0, 0)
	vp.SetContent("")

	// t := textinput.New()
	// t.CharLimit = 500
	// t.Placeholder = "Send a message"
	// t.Prompt = "▶ "
	// t.Focus()

	return Model{
		width:    width,
		height:   height,
		viewport: vp,
		// textinput: t,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)
	// m.textinput, tiCmd = m.textinput.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width - 2
		h := msg.Height - 7
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
		}
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	var b strings.Builder
	fmt.Println("KOSTAJ")
	// b.WriteString(m.labelBox.
	// 	SetWidth(m.viewport.Width).
	// 	RenderBoxWithTabs(m.chats, m.viewport.View()))
	// b.WriteString(fmt.Sprintf("%s ", m.renderRoom()))
	// b.WriteString(m.textinput.View())
	return b.String()
}

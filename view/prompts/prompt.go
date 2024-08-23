package prompts

import (
	"fmt"
	"strings"

	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/types"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/schollz/progressbar/v3"
)

type model struct {
	viewport       viewport.Model
	textinput      textinput.Model
	isEnterPressed bool
	twitch         *twitch.Client
	cfg            *types.JsonConfig
	promptedURL    string
	quality        twitch.VideoType
}

var (
	mainColor = lipgloss.Color("63")
	head      = fmt.Sprintf("%s\n\n", lipgloss.NewStyle().Background(mainColor).Render(" Twitch "))
)

func Open(twitch *twitch.Client, cfg *types.JsonConfig) {
	vp := viewport.New(0, 0)
	vp.SetContent("")

	t := textinput.New()
	t.CharLimit = 500
	t.Placeholder = ""
	t.Prompt = "Enter the Twitch link: "
	t.TextStyle = lipgloss.NewStyle().Faint(true)
	t.Focus()

	m := model{
		cfg:            cfg,
		twitch:         twitch,
		viewport:       vp,
		textinput:      t,
		isEnterPressed: false,
	}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		panic(err)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
	)
	m.textinput, tiCmd = m.textinput.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width
		h := msg.Height

		m.viewport.Width = w
		m.viewport.Height = h
		m.viewport.Style = lipgloss.
			NewStyle().
			Width(m.viewport.Width).
			Height(m.viewport.Height).
			Border(lipgloss.DoubleBorder()).
			Margin(1, 2).
			Padding(1, 2).
			BorderForeground(mainColor)

		m.viewport.SetContent(head + m.textinput.View())

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.isEnterPressed || m.textinput.Value() == "" {
				return m, tea.Batch(tiCmd)
			}
			m.handlePrompt()

		default:
			if !m.isEnterPressed {
				m.viewport.SetContent(head + m.textinput.View())
			}
		}
	}
	return m, tea.Batch(tiCmd)
}

func (m model) handlePrompt() (tea.Model, tea.Cmd) {
	msg := fmt.Sprintf("\nTwitch URL: %s\n", m.textinput.Value())

	bar := progressbar.DefaultBytes(-1, "Downloading: ")
	defer bar.Exit()

	if m.promptedURL != "" {
		m.promptedURL = m.textinput.Value()
	}

	// if m.quality != 3 {
	// 	id, vType, err := m.twitch.ID(m.promptedURL)
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return m, nil
	// 	}
	// 	if err := m.twitch.Downloader(id, vType, m.cfg.Paths.OutputPath, "best", 0, 0, bar); err != nil {
	// 		fmt.Println(err)
	// 		return m, nil
	// 	}
	// }

	v := lipgloss.
		NewStyle().
		Faint(true).
		Render(msg)
	m.viewport.SetContent(head + m.textinput.View() + v + bar.String())
	m.isEnterPressed = true
	m.textinput.Cursor.Blur()
	m.textinput.Reset()

	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString(m.viewport.View())
	return b.String()
}

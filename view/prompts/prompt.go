package prompts

import (
	"fmt"
	"time"

	qualities "github.com/Kostaaa1/twitchdl/internal"
	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/types"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type prompt struct {
	mediaURL     string
	mediaID      string
	mediaQuality string
	mediaVType   twitch.VideoType
	errorMsg     string
	start        time.Duration
	end          time.Duration
}

type model struct {
	twitch    *twitch.Client
	cfg       *types.JsonConfig
	msgChan   chan interface{}
	viewport  viewport.Model
	textinput textinput.Model
	prompt    *prompt
}

var (
	redColor    = lipgloss.Color("#C92D05")
	mainColor   = lipgloss.Color("63")
	promptStyle = lipgloss.NewStyle().Faint(true)
	errMsgStyle = lipgloss.NewStyle().Foreground(redColor)

	head         = fmt.Sprintf("%s\n\n", lipgloss.NewStyle().Bold(true).Background(mainColor).Render(" Twitch "))
	promptUrlMsg = "Enter the clip/video/stream URL: "
	qualityMsg   = "\nChoose the clip/video quality [best 1080 720 480 360 160 worst]: "
)

func Open(twitch *twitch.Client, cfg *types.JsonConfig) {
	vp := viewport.New(0, 0)
	vp.SetContent("")

	t := textinput.New()
	t.CharLimit = 500
	t.Placeholder = ""
	t.Prompt = promptUrlMsg
	t.Focus()

	m := model{
		cfg:       cfg,
		twitch:    twitch,
		viewport:  vp,
		textinput: t,
		msgChan:   make(chan interface{}),
		prompt:    &prompt{},
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
			Border(lipgloss.RoundedBorder()).
			Margin(1, 2).
			Padding(1, 2).
			BorderForeground(mainColor)
		m.viewport.SetContent(head + m.textinput.View())

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textinput.Value() == "" {
				return m, tea.Batch(tiCmd)
			}
			input := m.textinput.Value()
			m.prompt.errorMsg = ""
			m.textinput.SetValue("")

			if m.prompt.mediaURL == "" {
				id, vType, err := m.twitch.ID(input)
				if err != nil {
					m.prompt.errorMsg = err.Error()
					m.viewport.SetContent(m.renderView())
					return m, tea.Batch(tiCmd)
				}
				m.prompt.mediaID = id
				m.prompt.mediaVType = vType
				m.prompt.mediaURL = input
				m.textinput.Prompt = qualityMsg
				m.textinput.Focus()
			} else {
				if m.prompt.errorMsg == "" && qualities.IsQualityValid(input) {
					m.prompt.mediaQuality = input

					// bar := progressbar.DefaultBytes(-1, "Downloading: ")
					// defer bar.Exit()
					// if err := m.twitch.Downloader(m.prompt.mediaID, m.prompt.mediaVType, m.cfg.Paths.OutputPath, input, 0, 0, bar); err != nil {
					// 	fmt.Println(err)
					// 	return m, nil
					// }
					// m.prompt.mediaQuality = input

					return m, tea.Batch(tiCmd)
				} else {
					m.prompt.errorMsg = "\nwrong quality [best 1080 720 480 320 160 worst]."
				}
			}
		}
	}
	m.viewport.SetContent(m.renderView())
	return m, tea.Batch(tiCmd)
}

func (m *model) renderView() string {
	if m.prompt.mediaURL != "" && m.prompt.mediaQuality != "" {
		// m.textinput.Reset()
		// m.textinput.Prompt = promptUrlMsg
		m.prompt = &prompt{}
		return head
	}

	m.textinput.TextStyle = promptStyle
	m.textinput.Cursor.Blink = false
	view := head

	if m.prompt.mediaURL == "" {
		view += promptStyle.Render(m.prompt.mediaURL) + m.textinput.View()
	}
	if m.prompt.mediaURL != "" && m.prompt.mediaQuality == "" {
		view += promptUrlMsg + promptStyle.Render(m.prompt.mediaURL) + m.textinput.View()
	}
	if m.prompt.errorMsg != "" {
		view += errMsgStyle.Render("\n" + m.prompt.errorMsg)
	}
	return view
}

func (m model) View() string {
	return m.viewport.View()
}

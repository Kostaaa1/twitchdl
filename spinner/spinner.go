package spinner

import (
	"fmt"
	"os"
	"strings"

	"github.com/Kostaaa1/twitchdl/internal/utils"
	"github.com/Kostaaa1/twitchdl/types"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

type model struct {
	data         []types.ProgressBarState
	progressChan chan types.ProgressBarState
	spinner      spinner.Model
	quitting     bool
	err          error
}

func initialModel(titles []string, progChan chan types.ProgressBarState) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	var state []types.ProgressBarState
	for i := range titles {
		state = append(state, types.ProgressBarState{
			Text:      titles[i],
			ByteCount: 0,
		})
	}

	return model{
		spinner:      s,
		data:         state,
		progressChan: progChan,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.waitForMsg())
}

type chanMsg struct {
	Data types.ProgressBarState
}

func (m *model) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		return chanMsg{Data: <-m.progressChan}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		default:
			return m, nil
		}

	case errMsg:
		m.err = msg
		return m, nil

	case chanMsg:
		for i := range m.data {
			if m.data[i].Text == msg.Data.Text {
				if msg.Data.IsDone {
					m.data[i].IsDone = true
				}
				// m.data[i].ByteCount += msg.Data.ByteCount
				break
			}
		}
		return m, m.waitForMsg()

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	var str strings.Builder
	for i := range m.data {
		// this would be the message:
		// Downloading (22 MB, 4.802 MB/s) [1s]
		// this would be the speed:
		// s.KBsPerSecond = float64(p.state.currentBytes) / 1024.0 / s.SecondsSince

		var s string
		if m.data[i].IsDone {
			s += fmt.Sprintf("\n✅ %s: (%s) \n", m.data[i].Text, utils.ConvertBytes(m.data[i].ByteCount))
		} else {
			s += fmt.Sprintf("\n %s%s: (%s) \n", m.spinner.View(), m.data[i].Text, utils.ConvertBytes(m.data[i].ByteCount))
		}
		str.WriteString(s)
	}
	if m.quitting {
		return str.String() + "\n"
	}
	return str.String()
}

func Open(titles []string, progressChan chan types.ProgressBarState) {
	p := tea.NewProgram(initialModel(titles, progressChan))
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

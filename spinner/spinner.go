package spinner

import (
	"fmt"
	"os"
	"strings"
	"time"

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
			Text:        titles[i],
			IsDone:      false,
			ByteCount:   0,
			StartTime:   time.Now(),
			CurrentTime: 0,
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
				m.data[i].ByteCount += msg.Data.ByteCount
				m.data[i].CurrentTime = time.Since(m.data[i].StartTime).Seconds()
				// if m.data[i].CurrentTime > 0 {
				// m.data[i].KBsPerSecond = float64(m.data[i].ByteCount) / (1024.0 * 1024.0) / m.data[i].CurrentTime
				// }
				if msg.Data.IsDone {
					m.data[i].IsDone = true
				}
				break
			}
		}

		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd, m.waitForMsg())

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd, m.waitForMsg())
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	var str strings.Builder
	str.WriteString("\n")
	for i := 0; i < len(m.data); i++ {
		downloadMsg := fmt.Sprintf("(%s) [%.0fs]", utils.ConvertBytes(m.data[i].ByteCount), m.data[i].CurrentTime)
		if m.data[i].IsDone {
			s := fmt.Sprintf("✅ %s: %s \n", m.data[i].Text, downloadMsg)
			str.WriteString(s)
		} else {
			s := fmt.Sprintf(" %s%s: %s \n", m.spinner.View(), m.data[i].Text, downloadMsg)
			str.WriteString(s)
		}
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

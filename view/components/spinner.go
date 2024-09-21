package components

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/internal/bytecount"
	"github.com/Kostaaa1/twitchdl/types"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type errMsg error

type SpinnerState struct {
	text        string
	totalBytes  float64
	startTime   time.Time
	elapsedTime float64
	isDone      bool
	err         error
}

type model struct {
	state        []SpinnerState
	progressChan chan types.ProgresbarChanData
	spinner      spinner.Model
	quitting     bool
	err          error
}

func initialModel(titles []string, progChan chan types.ProgresbarChanData) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner:      s,
		state:        initSpinnerState(titles),
		progressChan: progChan,
	}
}

func initSpinnerState(titles []string) []SpinnerState {
	var state []SpinnerState
	for i := range titles {
		state = append(state, SpinnerState{
			text:        titles[i],
			totalBytes:  0,
			startTime:   time.Now(),
			elapsedTime: 0,
			isDone:      false,
		})
	}
	return state
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.waitForMsg())
}

type chanMsg struct {
	types.ProgresbarChanData
}

func (m *model) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		return chanMsg{ProgresbarChanData: <-m.progressChan}
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
		for i := range m.state {
			if m.state[i].text == msg.Text {
				m.state[i].totalBytes += float64(msg.Bytes)
				// m.state[i].CurrentTime = time.Since(m.state[i].StartTime).Seconds()
				// m.state[i].ByteCount.Convert()
				// if m.state[i].CurrentTime > 0 {
				// m.state[i].KBsPerSecond = float64(m.state[i].ByteCount) / (1024.0 * 1024.0) / m.state[i].CurrentTime
				// }

				if msg.IsDone {
					m.state[i].isDone = true
				}
				break
			}
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd, m.waitForMsg())

	default:
		var cmd tea.Cmd
		m.updateTime()
		m.spinner, cmd = m.spinner.Update(msg)
		return m, tea.Batch(cmd, m.waitForMsg())
	}
}

func (m *model) updateTime() {
	for i := range m.state {
		m.state[i].elapsedTime = time.Since(m.state[i].startTime).Seconds()
	}
}

func (m *model) getProgressMsg(total, ctime float64) string {
	b := bytecount.ConvertBytes(total)
	downloadMsg := fmt.Sprintf("(%.1f %s) [%.0fs]", b.Total, b.Unit, ctime)
	return downloadMsg
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	var str strings.Builder
	str.WriteString("\n")
	for i := 0; i < len(m.state); i++ {
		if m.state[i].err != nil {
			s := fmt.Sprintf("⚠️ Failed to download: %s \n", m.state[i].err)
			str.WriteString(s)
			continue
		}

		downloadMsg := m.getProgressMsg(m.state[i].totalBytes, m.state[i].elapsedTime)

		if m.state[i].isDone {
			s := fmt.Sprintf("✅ %s: %s \n", m.state[i].text, downloadMsg)
			str.WriteString(s)
		} else {
			s := fmt.Sprintf(" %s%s: %s \n", m.spinner.View(), m.state[i].text, downloadMsg)
			str.WriteString(s)
		}
	}
	return str.String()
}

func Spinner(titles []string, progressChan chan types.ProgresbarChanData) {
	p := tea.NewProgram(initialModel(titles, progressChan))
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

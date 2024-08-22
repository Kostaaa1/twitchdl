package root

import (
	"fmt"
	"os"
	"strings"

	"github.com/Kostaaa1/twitchdl/view/chat"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var mainColor = lipgloss.Color("63")
var docStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(mainColor)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			item := m.list.SelectedItem().FilterValue()
			if strings.HasPrefix(item, "Download") {
			}
			if strings.HasPrefix(item, "Record") {
			}
			if strings.HasPrefix(item, "Chats") {
				chat.Open()
			}
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func Open() {
	var items = []list.Item{
		item{title: "Download livestream", desc: "Provide URL of Clip or VOD to download."},
		item{title: "Record livestream", desc: "Record livestream."},
		item{title: "Chats", desc: "Open chats."},
	}

	d := list.NewDefaultDelegate()
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.Foreground(mainColor).BorderLeftForeground(mainColor)
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.Foreground(mainColor).BorderLeftForeground(mainColor)
	d.Styles.FilterMatch = d.Styles.FilterMatch.Underline(false)

	m := model{list: list.New(items, d, 0, 0)}
	m.list.Title = "Twitch"

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

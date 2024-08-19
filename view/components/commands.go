package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func RenderCommands(w, h int) string {
	return lipgloss.NewStyle().Width(w - 2).Height(h).
		Padding(1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("3")).
		Render(prepareCommands(w))
}

func prepareCommands(w int) string {
	type Command struct {
		cmdType string
		cmd     string
		help    string
	}
	var commands = []Command{
		{cmdType: "key", cmd: "ctrl+c/esc", help: "exit"},
		{cmdType: "key", cmd: "ctrl+→", help: "next chat"},
		{cmdType: "key", cmd: "ctrl+←", help: "prev chat"},
		{cmdType: "key", cmd: "shift+ctrl+→", help: "move chat forward"},
		{cmdType: "key", cmd: "shift+ctrl+←", help: "move chat backwards"},
		{cmdType: "key", cmd: "tab", help: "close commands window"},
		{cmdType: "key", cmd: "ctrl+o", help: "opens livestream in media player"},
		{cmdType: "key", cmd: "ctrl+i", help: "opens window with followed livestreams"},

		{cmdType: "input", cmd: "/add [channel]", help: "adds new chat tab"},
		{cmdType: "input", cmd: "/info [channel]", help: "info channel"},
		{cmdType: "input", cmd: "/record [stream url]", help: "record the livestream"},
		{cmdType: "input", cmd: "/download [vod/clip url]", help: "download the twitch clip or vod"},
	}

	var b strings.Builder
	for i, cmd := range commands {
		if i == 0 {
			b.WriteString("Key commands:\n\n")
		}
		if i == 5 {
			b.WriteString("\nInput commands:\n\n")
		}
		b.WriteString(fmt.Sprintf("%d) %s - %s\n", i+1, cmd.cmd, cmd.help))
	}
	return lipgloss.NewStyle().Width(w - 6).Faint(true).Render(b.String())
}

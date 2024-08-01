package chat

import (
	"fmt"
	"strings"

	"github.com/Kostaaa1/twitchdl/types"
	"github.com/Kostaaa1/twitchdl/utils"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

const (
	subColor          = "#04a5e5"
	announcementColor = "#40a02b"
	raidColor         = "#fe640b"
	firstMsgColor     = "#ea76db"
)

type BoxWithLabel struct {
	BoxStyle   lipgloss.Style
	LabelStyle lipgloss.Style
	width      int
	color      string
}

func NewBoxWithLabel(color string) BoxWithLabel {
	return BoxWithLabel{
		BoxStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(color)).
			Padding(0),
		LabelStyle: lipgloss.
			NewStyle().
			Padding(0),
		color: color,
	}
}

func (b *BoxWithLabel) SetWidth(width int) *BoxWithLabel {
	b.width = width
	return b
}

func (b *BoxWithLabel) renderLabel(chat *Chat, id int) string {
	border := lipgloss.Border{
		Top:         "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "│",
		BottomRight: "╰",
		Bottom:      "─",
	}

	if chat.IsActive {
		border.Bottom = " "
		if id == 0 {
			border.BottomLeft = "│"
		} else {
			border.BottomRight = "╰"
			border.BottomLeft = "╯"
		}
	} else {
		if id == 0 {
			border.BottomLeft = "├"
			border.BottomRight = "┴"
		} else {
			border.BottomRight = "┴"
			border.BottomLeft = "┴"
		}
	}

	l := b.LabelStyle.Border(border).
		BorderForeground(lipgloss.Color(b.color)).
		Bold(true).
		Italic(true).
		Padding(0)

	if chat.IsActive {
		l = l.Foreground(lipgloss.Color(b.color))
	}
	return l.Render(fmt.Sprintf(" %s ", chat.channelName))
}

func (b *BoxWithLabel) RenderBoxWithTabs(chats *[]Chat, content string) string {
	var (
		topBorderStyler func(strs ...string) string = lipgloss.NewStyle().
				Foreground(b.BoxStyle.GetBorderTopForeground()).
				Render
		border   lipgloss.Border = b.BoxStyle.GetBorderStyle()
		topLeft  string          = topBorderStyler(border.TopLeft)
		topRight string          = topBorderStyler(border.TopRight)
	)

	width := lipgloss.Width(content)
	if b.width != 0 {
		width = b.width
	}
	borderWidth := b.BoxStyle.GetHorizontalBorderSize()

	var stack []string
	for i := range *chats {
		stack = append(stack, b.renderLabel(&(*chats)[i], i))
	}
	// horLabels := lipgloss.JoinHorizontal(lipgloss.Position(0), renderedLabel, renderedLabel2)
	labels := lipgloss.JoinHorizontal(lipgloss.Position(0), stack...)
	cellsShort := max(0, width+borderWidth-lipgloss.Width(topLeft+topRight+labels))
	gap := strings.Repeat(border.Top, cellsShort+1)
	top := labels + topBorderStyler(gap) + topRight
	bottom := b.BoxStyle.
		BorderTop(false).
		Width(width).
		Render(content)
	return top + "\n" + bottom + "\n"
}

func (b *BoxWithLabel) RenderBox(label, content string) string {
	var (
		topBorderStyler func(strs ...string) string = lipgloss.NewStyle().
				Foreground(b.BoxStyle.GetBorderTopForeground()).
				Render
		border        lipgloss.Border = b.BoxStyle.GetBorderStyle()
		topLeft       string          = topBorderStyler(border.TopLeft)
		topRight      string          = topBorderStyler(border.TopRight)
		renderedLabel string          = b.LabelStyle.
				Bold(true).
				Italic(true).
				Padding(0).
				Render(label)
	)

	width := lipgloss.Width(content)
	if b.width != 0 {
		width = b.width
	}
	borderWidth := b.BoxStyle.GetHorizontalBorderSize()
	cellsShort := max(0, width+borderWidth-lipgloss.Width(topLeft+topRight+renderedLabel))
	gap := strings.Repeat(border.Top, cellsShort)
	top := topBorderStyler(border.TopLeft) + renderedLabel + topBorderStyler(gap) + topRight

	if width < lipgloss.Width(top) {
		content = content + strings.Repeat(" ", lipgloss.Width(top)-width-2)
		width = lipgloss.Width(top) - 2
	}

	bottom := b.BoxStyle.
		BorderTop(false).
		Width(width).
		Render(content)
	return top + "\n" + bottom
}

func usernameColorizer(color string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color))
}

// var (
// 	showBadges            bool
// 	showTimestamps        bool
// 	highlightSubs         bool
// 	highlightRaids        bool
// 	firstTimeChatterColor string
// 	watchedUsers          map[string]any
// )
// SetFormatterConfigValues sets the formatter customization options from the config.
// This is required because Viper won't have loaded the config yet when it reads this file.
// func SetFormatterConfigValues() {
// 	showBadges = viper.GetBool(ShowBadgesKey)
// 	showTimestamps = viper.GetBool(ShowTimestampsKey)
// 	highlightSubs = viper.GetBool(HighlightRaidsKey)
// 	highlightRaids = viper.GetBool(HighlightRaidsKey)
// 	watchedUsers = viper.GetStringMap(WatchedUsersKey)
// 	firstTimeChatterColor = viper.GetString(FirstTimeChatterColorKey)
// }

func GenerateIcon(userType string) string {
	switch userType {
	case "broadcaster":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#d20f39")).Render("[]")
	case "mod":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#40a02b")).Render("[⛨]")
	case "vip":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ea76cb")).Render("[★]")
	case "staff":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#8839ef")).Render("[★]")
	}
	return ""
}

// func formatMessageTimestamp(timestamp string, msg string) string {
// 	msgHeight := lipgloss.Height(msg)
// 	var newT string = timestamp
// 	for i := 1; i < msgHeight; i++ {
// 		newT += "\n" + strings.Repeat(" ", lipgloss.Width(timestamp))
// 	}
// 	return lipgloss.JoinHorizontal(1, newT, msg)
// }

func FormatChatMessage(message types.ChatMessage, width int) string {
	icon := GenerateIcon(message.Metadata.UserType)
	if message.Metadata.Color == "" {
		message.Metadata.Color = utils.GetRandHex()
	}
	msg := fmt.Sprintf(
		"%s%s: %s",
		icon,
		usernameColorizer(message.Metadata.Color).Render(message.Metadata.DisplayName),
		message.Message,
	)

	msg = wordwrap.String(msg, width-14)
	if !message.IsFirstMessage {
		timestamp := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("[%s] ", message.Metadata.Timestamp))
		// return formatMessageTimestamp(timestamp, msg)
		return fmt.Sprintf("%s%s", timestamp, msg)
	} else {
		box := NewBoxWithLabel(firstMsgColor)
		return box.RenderBox(" First message ", msg)
	}
}

func FormatSubMessage(message types.SubNotice, width int) string {
	icon := GenerateIcon(message.Metadata.UserType)
	if message.Metadata.Color == "" {
		message.Metadata.Color = utils.GetRandHex()
	}
	msg := fmt.Sprintf(
		"%s%s: ✯ %s",
		icon,
		usernameColorizer(message.Metadata.Color).Render(message.Metadata.DisplayName),
		message.Metadata.SystemMsg,
	)
	box := NewBoxWithLabel(subColor)
	msg = wordwrap.String(msg, width-20)
	label := lipgloss.NewStyle().Foreground(lipgloss.Color(subColor)).Render(fmt.Sprintf(" %s ", utils.Capitalize(message.SubPlan)))
	return box.RenderBox(label, msg)
}

// func FormatAnnouncementMessage(message types.AnnouncementMessage, width int) string {
// 	box := NewBoxWithLabel(announcementColor)
// 	msg := fmt.Sprintf(
// 		"%s: %s",
// 		usernameColorizer(message.Color).Render(message.DisplayName),
// 		message.Message,
// 	)
// 	msg = wordwrap.String(msg, width)
// 	return box.Render("Announcement", msg)
// }

// func FormatRaidMessage(message types.RaidMessage, width int) string {
// 	box := NewBoxWithLabel(raidColor)
// 	msg := fmt.Sprintf(
// 		"%s raided the channel with %s viewers!",
// 		usernameColorizer(message.Color).Render(message.DisplayName),
// 		message.ViewerCount,
// 	)
// 	msg = wordwrap.String(msg, width)
// 	if highlightRaids {
// 		return box.Render("Raid", msg)
// 	}
// 	return msg + "\n"
// }

// func FormatGiftSubMessage(message types.SubGiftMessage, width int) string {
// 	box := NewBoxWithLabel(subColor)
// 	msg := fmt.Sprintf(
// 		"%s gifted a subscription to %s!",
// 		usernameColorizer(message.Color).Render(message.GiverName),
// 		message.ReceiverName,
// 	)
// 	msg = wordwrap.String(msg, width)
// 	if highlightSubs {
// 		return box.Render("Gift sub", msg)
// 	}
// 	return msg + "\n"
// }

// func FormatMysteryGiftSubMessage(message types.MysterySubGiftMessage, width int) string {
// 	box := NewBoxWithLabel(subColor)
// 	msg := fmt.Sprintf(
// 		"%s is giving %s subs to the channel!",
// 		usernameColorizer(message.Color).Render(message.GiverName),
// 		message.GiftAmount,
// 	)
// 	msg = wordwrap.String(msg, width)
// 	if highlightSubs {
// 		return box.Render("Gifting subs", msg)
// 	}
// 	return msg + "\n"
// }

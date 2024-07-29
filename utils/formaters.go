package utils

import (
	"fmt"
	"strings"

	"github.com/Kostaaa1/twitchdl/types"
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

func (b *BoxWithLabel) RenderBoxWithTabs(label, content string) string {
	var (
		topBorderStyler func(strs ...string) string = lipgloss.NewStyle().
				Foreground(b.BoxStyle.GetBorderTopForeground()).
				Render
		border        lipgloss.Border = b.BoxStyle.GetBorderStyle()
		topLeft       string          = topBorderStyler(border.TopLeft)
		topRight      string          = topBorderStyler(border.TopRight)
		renderedLabel string          = b.LabelStyle.Border(lipgloss.Border{
			Top:         "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "╭",
			TopRight:    "╮",
			BottomLeft:  "│",
			BottomRight: "╰",
		}).
			BorderForeground(lipgloss.Color(b.color)).
			Bold(true).
			Italic(true).
			Padding(0).
			Render(label)
		// renderedLabel2 string = b.LabelStyle.Border(lipgloss.Border{
		// 	Top:         "─",
		// 	Bottom:      "─",
		// 	Left:        "│",
		// 	Right:       "│",
		// 	TopLeft:     "╭",
		// 	TopRight:    "╮",
		// 	BottomLeft:  "┴",
		// 	BottomRight: "┴",
		// }).
		// 	BorderForeground(lipgloss.Color(b.color)).
		// 	Padding(0).
		// 	Render(label)
	)

	width := lipgloss.Width(content)
	if b.width != 0 {
		width = b.width
	}

	borderWidth := b.BoxStyle.GetHorizontalBorderSize()
	horLabels := lipgloss.JoinHorizontal(lipgloss.Position(0), renderedLabel)
	cellsShort := max(0, width+borderWidth-lipgloss.Width(topLeft+topRight+horLabels))
	gap := strings.Repeat(border.Top, cellsShort+1)
	top := horLabels + topBorderStyler(gap) + topRight
	bottom := b.BoxStyle.
		BorderTop(false).
		Width(width).
		Render(content)
	return top + "\n" + bottom + "\n"
}

func (b *BoxWithLabel) Render(label, content string) string {
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
	cellsShort := max(0, width+borderWidth-lipgloss.Width(topLeft+topRight+renderedLabel)-1)
	gap := strings.Repeat(border.Top, cellsShort)
	top := topBorderStyler(border.TopLeft) + topBorderStyler(border.Top) + renderedLabel + topBorderStyler(gap) + topRight

	bottom := b.BoxStyle.
		BorderTop(false).
		Width(width).
		Render(content)
	return top + "\n" + bottom
}

func usernameColorizer(color string) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color))
}

var (
	showBadges            bool
	showTimestamps        bool
	highlightSubs         bool
	highlightRaids        bool
	firstTimeChatterColor string
	watchedUsers          map[string]any
)

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

// GenerateIcon returns a colored user-type icon, if applicable to the user.
// For example, a green sword icon for a moderator.
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

func FormatChatMessage(message types.ChatMessage, width int) string {
	// This is used:
	////////////////////////
	icon := GenerateIcon(message.Metadata.UserType)
	if message.Metadata.Color == "" {
		message.Metadata.Color = GetRandHex()
	}
	timestamp := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("[%s] ", message.Metadata.Timestamp))
	msg := fmt.Sprintf(
		"%s%s%s: %s",
		timestamp,
		icon,
		usernameColorizer(message.Metadata.Color).Render(message.Metadata.DisplayName),
		message.Message,
	)

	////////////////////////
	// msg = wordwrap.String(msg, width)
	// return msg
	// commented for dev:
	// msg = wordwrap.String(msg, width)
	////////////////////////

	if !message.IsFirstMessage {
		return msg
	}
	box := NewBoxWithLabel(firstMsgColor)
	if message.IsFirstMessage {
		msg = wordwrap.String(msg, width-20)
		return box.Render(" First message ", msg)
	}
	return msg
	////////////////////////
	return ""
}

// func FormatSubMessage(message types.SubMessage, width int) string {
// 	var fullMessage string
// 	if message.Message != "" {
// 		fullMessage = ": " + message.Message
// 	} else {
// 		fullMessage = "!"
// 	}
// 	box := NewBoxWithLabel(subColor)
// 	msg := fmt.Sprintf(
// 		"%s subscribed for %s months%s",
// 		usernameColorizer(message.Color).Render(message.DisplayName),
// 		message.Months,
// 		fullMessage,
// 	)
// 	msg = wordwrap.String(msg, width)
// 	if highlightSubs {
// 		return box.Render("Sub", msg)
// 	}
// 	return msg + "\n"
// }

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

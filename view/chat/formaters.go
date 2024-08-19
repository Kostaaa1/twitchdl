package chat

import (
	"fmt"

	"github.com/Kostaaa1/twitchdl/types"
	"github.com/Kostaaa1/twitchdl/utils"
	"github.com/Kostaaa1/twitchdl/view/components"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

const (
	subColor = "#04a5e5"
	// announcementColor = "#40a02b"
	raidColor     = "#fe640b"
	firstMsgColor = "#ea76db"
)

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
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#d20f39")).Render(" [] ")
	case "mod":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#40a02b")).Render(" [⛨] ")
	case "vip":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ea76cb")).Render(" [★] ")
	case "staff":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#8839ef")).Render(" [★] ")
	}
	return " "
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
	if !message.Metadata.IsFirstMessage {
		timestamp := lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("[%s]", message.Metadata.Timestamp))
		return fmt.Sprintf("%s%s", timestamp, msg)
	} else {
		box := components.NewBoxWithLabel(firstMsgColor)
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
	box := components.NewBoxWithLabel(subColor)
	msg = wordwrap.String(msg, width-50)
	label := lipgloss.NewStyle().Foreground(lipgloss.Color(subColor)).Render(fmt.Sprintf(" %s ", utils.Capitalize(message.SubPlan)))
	return box.RenderBox(label, msg)
}

func FormatRaidMessage(message types.RaidNotice, width int) string {
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
	box := components.NewBoxWithLabel(raidColor)
	msg = wordwrap.String(msg, width-50)
	label := lipgloss.NewStyle().Foreground(lipgloss.Color(raidColor)).Render("Raid")
	return box.RenderBox(label, msg)
}

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

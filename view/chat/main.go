package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/Kostaaa1/twitchdl/internal/utils"
	"github.com/Kostaaa1/twitchdl/twitch"
	"github.com/Kostaaa1/twitchdl/types"
	"github.com/Kostaaa1/twitchdl/view/components"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

type NewChannelMessage struct {
	Data interface{}
}

type Model struct {
	twitch              *twitch.Client
	ws                  *WebSocketClient
	viewport            viewport.Model
	labelBox            components.BoxWithLabel
	textinput           textinput.Model
	width               int
	height              int
	msgChan             chan interface{}
	chats               []types.Chat
	showCommands        bool
	commandsWindowWidth int
	err                 error
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

func Open(twitch *twitch.Client, cfg *types.JsonConfig) {
	vp := viewport.New(0, 0)
	vp.SetContent("")
	t := textinput.New()
	t.CharLimit = 500
	t.Placeholder = "Send a message"
	t.Prompt = " ▶ "
	t.Focus()

	msgChan := make(chan interface{})
	ws, err := CreateWSClient()
	if err != nil {
		panic(err)
	}

	go func() {
		if err := ws.Connect(cfg.Creds.AccessToken, cfg.Creds.ClientID, msgChan, cfg.OpenedChats); err != nil {
			fmt.Println("Connection error: ", err)
		}
	}()

	chats := []types.Chat{}
	for i, channel := range cfg.OpenedChats {
		chats = append(chats, createNewChat(channel, i == 0))
	}

	m := Model{
		twitch:              twitch,
		ws:                  ws,
		chats:               chats,
		err:                 nil,
		width:               0,
		height:              0,
		msgChan:             msgChan,
		labelBox:            components.NewBoxWithLabel(cfg.Colors.Primary),
		viewport:            vp,
		textinput:           t,
		showCommands:        false,
		commandsWindowWidth: 32,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func createNewChat(channel string, isActive bool) types.Chat {
	return types.Chat{
		IsActive: isActive,
		Messages: []string{
			lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("Welcome to %s channel", channel)),
		},
		Room:    types.Room{},
		Channel: channel,
	}
}

func (m Model) Init() tea.Cmd {
	return m.waitForMsg()
}

var errTimer *time.Timer

func (m *Model) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		newMsg := <-m.msgChan
		switch newMsg.(type) {
		case errMsg:
			if errTimer != nil {
				errTimer.Stop()
			}
			errTimer = time.AfterFunc(time.Second*2, func() {
				m.msgChan <- errMsg{err: nil}
			})
			return newMsg
		default:
			return NewChannelMessage{Data: newMsg}
		}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
	)
	m.textinput, tiCmd = m.textinput.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width - 2
		h := msg.Height - 8
		m.labelBox.SetWidth(w)
		m.viewport.Width = w
		m.viewport.Height = h
		m.width = w
		m.height = h
		m.viewport.Style = lipgloss.
			NewStyle().
			Width(m.viewport.Width).
			Height(m.viewport.Height)

		if m.chats[0].IsActive {
			m.updateChatViewport(&m.chats[0])
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			viper.WriteConfig()
			return m, tea.Quit

		case tea.KeyEnter:
			m.sendMessage()
		case tea.KeyCtrlRight:
			m.nextTab()
		case tea.KeyCtrlLeft:
			m.prevTab()
		case tea.KeyCtrlShiftRight:
			m.moveTabForward()
		case tea.KeyCtrlShiftLeft:
			m.moveTabBack()
		case tea.KeyCtrlW:
			if len(m.chats) > 1 {
				m.removeActiveChat()
			}
		case tea.KeyCtrlO:
			go func() {
				chat := m.getActiveChat()
				if err := m.twitch.OpenStreamInMediaPlayer(chat.Channel); err != nil {
					m.msgChan <- errMsg{err: err}
				}
			}()
		case tea.KeyTab:
			m.showCommands = !m.showCommands
			if m.showCommands {
				m.viewport.Width = m.width - m.commandsWindowWidth
			} else {
				m.viewport.Width = m.width
			}
		}

	case errMsg:
		m.err = msg.err
		return m, m.waitForMsg()

	case NewChannelMessage:
		switch chanMsg := msg.Data.(type) {
		case types.Room:
			m.addRoomToChat(chanMsg)
		case types.ChatMessage:
			chat := m.getChat(chanMsg.Metadata.RoomID)
			if chat != nil {
				m.appendMessage(chat, FormatChatMessage(chanMsg, m.width))
			}
		case types.SubNotice:
			chat := m.getChat(chanMsg.Metadata.RoomID)
			if chat != nil {
				m.appendMessage(chat, FormatSubMessage(chanMsg, m.width))
			}
		case types.Notice:
			chat := m.getChat(chanMsg.DisplayName)
			if chat != nil {
				m.appendMessage(chat, chanMsg.SystemMsg)
			}
		}
		return m, m.waitForMsg()
	}

	return m, tea.Batch(tiCmd)
}

func (m Model) View() string {
	var b strings.Builder
	main := m.labelBox.
		SetWidth(m.viewport.Width).
		RenderBoxWithTabs(m.chats, m.viewport.View())

	if !m.showCommands {
		b.WriteString(main)
	} else {
		b.WriteString(lipgloss.
			JoinHorizontal(lipgloss.Position(0.5), main, components.RenderCommands(m.commandsWindowWidth, m.height)))
	}
	b.WriteString("\n" + lipgloss.JoinHorizontal(lipgloss.Position(0), m.renderRoomState(), m.textinput.View()))
	b.WriteString(m.renderError())
	return b.String()
}

func (m *Model) createNewMessage(chat *types.Chat) types.ChatMessage {
	newMessage := types.ChatMessage{
		Message: m.textinput.Value(),
		Metadata: types.ChatMessageMetadata{
			Metadata: types.Metadata{
				Color:        chat.Room.Metadata.Color,
				DisplayName:  chat.Room.Metadata.DisplayName,
				IsMod:        chat.Room.Metadata.IsMod,
				IsSubscriber: chat.Room.Metadata.IsSubscriber,
				UserType:     chat.Room.Metadata.UserType,
			},
			RoomID:    chat.Room.RoomID,
			Timestamp: utils.GetCurrentTimeFormatted(),
		},
	}
	return newMessage
}

func (m *Model) renderError() string {
	var b strings.Builder
	if m.err != nil {
		b.WriteString(lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("\n\n[NOTIFY] - %s", m.err)))
	} else {
		b.WriteString("")
	}
	return b.String()
}

func (m *Model) sendMessage() {
	if m.textinput.Value() == "" {
		return
	}
	input := m.textinput.Value()
	if !strings.HasPrefix(input, "/") {
		chat := m.getActiveChat()
		newMessage := m.createNewMessage(chat)
		m.ws.FormatIRCMsgAndSend("PRIVMSG", chat.Channel, input)
		chat.Messages = append(chat.Messages, FormatChatMessage(newMessage, m.width))
		m.updateChatViewport(chat)
	} else {
		m.handleInputCommand(input)
	}
	m.textinput.Reset()
}

func (m *Model) handleInputCommand(cmd string) {
	parts := strings.Split(cmd, " ")
	if len(parts) > 2 {
		return
	}
	switch parts[0] {
	case "/add":
		m.addChat(parts[1])
	case "/info":
		fmt.Println(parts[1])
	default:
		m.msgChan <- errMsg{err: fmt.Errorf("invalid command: %s", cmd)}
	}
}

func (m *Model) addChat(channelName string) {
	newChat := createNewChat(channelName, false)
	m.chats = append(m.chats, newChat)
	m.ws.ConnectToChannel(newChat.Channel)
	newChannels := []string{}
	for _, c := range m.chats {
		newChannels = append(newChannels, c.Channel)
	}
	viper.Set("openedChats", newChannels)
	// viper.WriteConfig()
}

func (m *Model) addRoomToChat(chanMsg types.Room) {
	for i := range m.chats {
		c := &(m.chats)[i]
		if c.Channel == chanMsg.Metadata.Channel {
			c.Room = chanMsg
			break
		}
	}
}

func (m *Model) removeActiveChat() {
	var activeChan string
	chats := []types.Chat{}
	newActiveId := 0

	for i, c := range m.chats {
		if !c.IsActive {
			chats = append(chats, c)
		} else {
			activeChan = c.Channel
			m.ws.LeaveChannel(c.Channel)
			newActiveId = i
			if i == len(m.chats)-1 {
				newActiveId--
			}
		}
	}
	chats[newActiveId].IsActive = true
	newActiveC := chats[newActiveId]
	m.updateChatViewport(&newActiveC)

	// remove from config...
	activeChans := viper.GetStringSlice("openedChats")
	newOpenedChats := []string{}
	for _, ch := range activeChans {
		if ch != activeChan {
			newOpenedChats = append(newOpenedChats, ch)
		}
	}
	viper.Set("openedChats", newOpenedChats)
	// viper.WriteConfig()
	m.chats = chats
}

func (m *Model) appendMessage(chat *types.Chat, message string) {
	if len(chat.Messages) > 100 {
		chat.Messages = chat.Messages[1:]
	}
	chat.Messages = append(chat.Messages, message)
	if chat.IsActive {
		m.updateChatViewport(chat)
	}
}

func (m *Model) updateChatViewport(chat *types.Chat) {
	m.viewport.SetContent(strings.Join(chat.Messages, "\n"))
	m.viewport.GotoBottom()
}

// TODO :
func (m *Model) moveTabForward() {
	openedChats := make([]string, len(m.chats))
	for i := len(m.chats) - 1; i >= 0; i-- {
		if i > 0 && m.chats[i-1].IsActive {
			m.chats[i], m.chats[i-1] = m.chats[i-1], m.chats[i]
		}
		openedChats[i] = m.chats[i].Channel
	}
	viper.Set("openedChats", openedChats)
}

func (m *Model) moveTabBack() {
	openedChats := make([]string, len(m.chats))
	for i := range m.chats {
		if i < len(m.chats)-1 && m.chats[i+1].IsActive {
			m.chats[i], m.chats[i+1] = m.chats[i+1], m.chats[i]
		}
		openedChats[i] = m.chats[i].Channel
	}
	viper.Set("openedChats", openedChats)
}

func (m *Model) nextTab() {
	var activeIndex int
	for i, chat := range m.chats {
		if chat.IsActive {
			activeIndex = i
			break
		}
	}
	(m.chats)[activeIndex].IsActive = false
	nextIndex := (activeIndex + 1) % len(m.chats)
	(m.chats)[nextIndex].IsActive = true
	m.updateChatViewport(&(m.chats)[nextIndex])
}

func (m *Model) prevTab() {
	var activeIndex int
	for i, c := range m.chats {
		if c.IsActive {
			activeIndex = i
			break
		}
	}
	(m.chats)[activeIndex].IsActive = false
	prevIndex := (activeIndex - 1 + len(m.chats)) % len(m.chats)
	(m.chats)[prevIndex].IsActive = true
	m.updateChatViewport(&(m.chats)[prevIndex])
}

func (m Model) getActiveChat() *types.Chat {
	for i := range m.chats {
		if (m.chats)[i].IsActive {
			return &(m.chats[i])
		}
	}
	return nil
}

func (m Model) getChat(roomID string) *types.Chat {
	for i := range m.chats {
		if (m.chats)[i].Room.RoomID == roomID || (m.chats)[i].Channel == roomID {
			return &(m.chats[i])
		}
	}
	return nil
}

func (m Model) renderRoomState() string {
	chat := m.getActiveChat()
	style := lipgloss.NewStyle().Faint(true)
	switch {
	case chat.Room.IsEmoteOnly:
		return style.Render("[Emote-Only Chat]")
	case chat.Room.IsFollowersOnly:
		return style.Render("[Followers-Only Chat]")
	case chat.Room.IsSubsOnly:
		return style.Render("[Subscriber-Only Chat]")
	default:
		return ""
	}
}

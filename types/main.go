package types

type UserIRC struct {
	Message        string
	Badges         []string
	Color          string
	DisplayName    string
	IsFirstMessage bool
	IsMod          bool
	IsSubscriber   bool
	ID             string
	Timestamp      string
	Type           string
}

type RoomState struct {
	Color           string
	DisplayName     string
	IsMod           bool
	IsSubscriber    bool
	UserType        string
	IsEmoteOnly     bool
	IsFollowersOnly bool
	RoomID          string
	IsSubsOnly      bool
}

// msg-param-cumulative-months
// msg-param-goal-current-contributions=24577
// type UserNoticeMsg struct {
// 	UserIRC
// 	MsgID string
// }

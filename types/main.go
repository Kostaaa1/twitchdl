package types

type Metadata struct {
	Color        string
	DisplayName  string
	IsMod        bool
	IsSubscriber bool
	UserType     string
}

type ChatMessageMetadata struct {
	Metadata
	RoomID    string
	Timestamp string
	UserType  string
}

type ChatMessage struct {
	Metadata       ChatMessageMetadata
	Message        string
	IsFirstMessage bool
}

type RoomStateMetadata struct {
	Metadata
}

type RoomState struct {
	Metadata        RoomStateMetadata
	RoomID          string
	IsEmoteOnly     bool
	IsFollowersOnly bool
	IsSubsOnly      bool
}

type NoticeMetadata struct {
	Metadata
	MsgID     string
	RoomID    string
	SystemMsg string
	Timestamp string
	UserID    string
}

type RaidNotice struct {
	Metadata         NoticeMetadata
	ParamDisplayName string
	ParamLogin       string
	ViewerCount      int
}

type SubGiftNotice struct {
	Metadata             NoticeMetadata
	Months               int
	RecipientDisplayName string
	RecipientID          string
	RecipientName        string
	SubPlan              string
}

// type SubNotice struct {
// 	Metadata         NoticeMetadata
// 	CumulativeMonths int
// 	Months           int
// 	SubPlan          string
// }

type ResubNotice struct {
	Metadata         NoticeMetadata
	CumulativeMonths int
	Months           int
	SubPlan          string
	WasGifted        bool
}

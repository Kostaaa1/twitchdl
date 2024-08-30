package types

import (
	"time"
)

type Metadata struct {
	Color        string
	DisplayName  string
	IsMod        bool
	IsSubscriber bool
	UserType     string
}

type ChatMessageMetadata struct {
	Metadata
	RoomID         string
	IsFirstMessage bool
	Timestamp      string
}

type ChatMessage struct {
	Metadata ChatMessageMetadata
	Message  string
}

type RoomMetadata struct {
	Metadata
	Channel string
}

type Room struct {
	Metadata        RoomMetadata
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

type SubNotice struct {
	Metadata  NoticeMetadata
	Months    int
	SubPlan   string
	WasGifted bool
}

type Notice struct {
	MsgID       string
	DisplayName string
	SystemMsg   string
}

type JsonConfig struct {
	OpenedChats     []string  `json:"openedChats"`
	BroadcasterType string    `json:"broadcasterType"`
	Colors          Colors    `json:"colors"`
	CreatedAt       time.Time `json:"createdAt"`
	Creds           struct {
		AccessToken  string `json:"accessToken"`
		ClientID     string `json:"clientID"`
		ClientSecret string `json:"clientSecret"`
	} `json:"creds"`
	Description     string `json:"description"`
	DisplayName     string `json:"displayName"`
	ID              string `json:"id"`
	Login           string `json:"login"`
	OfflineImageUrl string `json:"offlineImageUrl"`
	Paths           struct {
		ChromePath string `json:"chromePath"`
		OutputPath string `json:"outputPath"`
	} `json:"paths"`
	ProfileImageUrl string `json:"profileImageUrl"`
	ShowTimestamps  bool   `json:"showTimestamps"`
	Type            string `json:"type"`
}

type Colors struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
	Danger    string `json:"danger"`
	Border    string `json:"border"`
	Icons     struct {
		Broadcaster string `json:"broadcaster"`
		Mod         string `json:"mod"`
		Staff       string `json:"staff"`
		Vip         string `json:"vip"`
	} `json:"icons"`
	Messages struct {
		Announcement string `json:"announcement"`
		First        string `json:"first"`
		Original     string `json:"original"`
		Raid         string `json:"raid"`
		Sub          string `json:"sub"`
	} `json:"messages"`
	Timestamp string `json:"timestamp"`
}

type SpinnerState struct {
	Text        string
	ByteCount   float64
	IsDone      bool
	StartTime   time.Time
	CurrentTime float64
}

type ProgresbarChanData struct {
	Text   string
	Bytes  int64
	IsDone bool
}

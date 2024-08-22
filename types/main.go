package types

import "time"

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
	ID              string    `json:"id"`
	Login           string    `json:"login"`
	DisplayName     string    `json:"displayName"`
	Type            string    `json:"type"`
	BroadcasterType string    `json:"broadcasterType"`
	Description     string    `json:"description"`
	ProfileImageURL string    `json:"profileImageUrl"`
	OfflineImageURL string    `json:"offlineImageUrl"`
	CreatedAt       time.Time `json:"createdAt"`
	ActiveChannels  []string  `json:"activeChannels"`
	Creds           struct {
		AccessToken  string `json:"accessToken"`
		ClientID     string `json:"clientId"`
		ClientSecret string `json:"clientSecret"`
	} `json:"creds"`
	ShowTimestamps bool `json:"showTimestamps"`
	Colors         struct {
		MainWindowBackground string `json:"mainWindowBackground"`
		Border               string `json:"border"`
		Timestamp            string `json:"timestamp"`
		Messages             struct {
			Original string `json:"original"`
			Raid     string `json:"raid"`
			Sub      string `json:"sub"`
			First    string `json:"first"`
			Subgif   string `json:"subgif"`
		} `json:"messages:"`
	} `json:"colors"`
	Paths struct {
		ChromePath string `json:"chromePath"`
		OutputPath string `json:"outputPath"`
	} `json:"paths"`
}

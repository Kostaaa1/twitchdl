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
	OpenedChats     []string  `json:"openedChats"`
	Broadcastertype string    `json:"broadcastertype"`
	Colors          Colors    `json:"colors"`
	Createdat       time.Time `json:"createdat"`
	Creds           struct {
		AccessToken  string `json:"accesstoken"`
		ClientID     string `json:"clientid"`
		ClientSecret string `json:"clientsecret"`
	} `json:"creds"`
	Description     string   `json:"description"`
	Displayname     string   `json:"displayname"`
	ID              string   `json:"id"`
	Login           string   `json:"login"`
	Offlineimageurl string   `json:"offlineimageurl"`
	Openedchats     []string `json:"openedchats"`
	Paths           struct {
		Chromepath string `json:"chromepath"`
		Outputpath string `json:"outputpath"`
	} `json:"paths"`
	Profileimageurl string `json:"profileimageurl"`
	Showtimestamps  bool   `json:"showtimestamps"`
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
		Announcmentcolor string `json:"announcmentcolor"`
		First            string `json:"first"`
		Original         string `json:"original"`
		Raid             string `json:"raid"`
		Subgif           string `json:"subgif"`
	} `json:"messages"`
	Timestamp string `json:"timestamp"`
}

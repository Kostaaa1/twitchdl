package main

import "github.com/Kostaaa1/twitchdl/chat"

var (
	name        = "slorpglorpski"
	accessToken = "x1ug4nduxyhopsdc1zrwbi1c3f5m0f"
	clientID    = "z4qytet5kietgqy0q7nxrgr8sverf1"
	secret      = "dzsedgplhczx5n0k25oj04339q0wei"
)

const (
	// viewmilestone = `@badge-info=;badges=bits/100;color=#8A2BE2;display-name=HatefulWordss;emotes=;flags=;id=fa06d7c2-c724-4d14-beac-5a220e6a9bd5;login=hatefulwordss;mod=0;msg-id=viewermilestone;msg-param-category=watch-streak;msg-param-copoReward=350;msg-param-id=cb2875b1-4780-417a-9ddd-12a6edc79e9c;msg-param-value=3;room-id=151368796;subscriber=0;system-msg=HatefulWordss\swatched\s3\sconsecutive\sstreams\sthis\smonth\sand\ssparked\sa\swatch\sstreak!;tmi-sent-ts=1722101522165;user-id=496917062;user-type=;vip=0 :tmi.twitch.tv USERNOTICE #piratesoftware :youve become my go to stream now because of how long your stream LOL no longer have to stop every so often to find someone else to listen t`

	subMessage = `@badge-info=subscriber/1;badges=subscriber/0,premium/1;color=#9ACD32;display-name=RannDC;emotes=;flags=;id=250b73de-5b32-451a-8eb6-a8ceaeeb7d12;login=ranndc;mod=0;msg-id=sub;msg-param-cumulative-months=1;msg-param-goal-contribution-type=SUB_POINTS;msg-param-goal-current-contributions=38290;msg-param-goal-target-contributions=77777;msg-param-goal-user-contributions=1;msg-param-months=0;msg-param-multimonth-duration=1;msg-param-multimonth-tenure=0;msg-param-should-share-streak=0;msg-param-sub-plan-name=Download\sA\sPirate;msg-param-sub-plan=Prime;msg-param-was-gifted=false;room-id=151368796;subscriber=1;system-msg=RannDC\ssubscribed\swith\sPrime.;tmi-sent-ts=1722100427940;user-id=21385513;user-type=;vip=0 :tmi.twitch.tv USERNOTICE #piratesoftware`

	resub = `@badge-info=subscriber/4;badges=subscriber/3,premium/1;color=#8A2BE2;display-name=MrSykez;emotes=;flags=;id=6eb598eb-2a53-45cd-a3ba-ee11efad754f;login=mrsykez;mod=0;msg-id=resub;msg-param-cumulative-months=4;msg-param-months=0;msg-param-multimonth-duration=1;msg-param-multimonth-tenure=0;msg-param-should-share-streak=0;msg-param-sub-plan-name=Größter\sFehler.\s(papaplatte);msg-param-sub-plan=Prime;msg-param-was-gifted=false;room-id=50985620;subscriber=1;system-msg=MrSykez\ssubscribed\swith\sPrime.\sThey've\ssubscribed\sfor\s4\smonths!;tmi-sent-ts=1722181131806;user-id=59804732;user-type=;vip=0 :tmi.twitch.tv USERNOTICE #papaplatte`

	subgift = `@badge-info=;badges=staff/1,premium/1;color=#0000FF;display-name=TWW2;emotes=;id=e9176cd8-5e22-4684-ad40-ce53c2561c5e;login=tww2;mod=0;msg-id=subgift;msg-param-months=1;msg-param-recipient-display-name=Mr_Woodchuck;msg-param-recipient-id=55554444;msg-param-recipient-name=mr_woodchuck;msg-param-sub-plan-name=House\sof\sNyoro~n;msg-param-sub-plan=1000;room-id=19571752;subscriber=0;system-msg=TWW2\sgifted\sa\sTier\s1\ssub\sto\sMr_Woodchuck!;tmi-sent-ts=1521159445153;turbo=0;user-id=87654321;user-type=staff :tmi.twitch.tv USERNOTICE #forstycup`

	raid = `@badge-info=;badges=turbo/1;color=#9ACD32;display-name=TestChannel;emotes=;id=3d830f12-795c-447d-af3c-ea05e40fbddb;login=testchannel;mod=0;msg-id=raid;msg-param-displayName=TestChannel;msg-param-login=testchannel;msg-param-viewerCount=15;room-id=33332222;subscriber=0;system-msg=15\sraiders\sfrom\sTestChannel\shave\sjoined\n!;tmi-sent-ts=1507246572675;turbo=1;user-id=123456;user-type= :tmi.twitch.tv USERNOTICE #othertestchannel`

	prime = `USERNOTICE MSG:  @badge-info=subscriber/1;badges=subscriber/0,premium/1;color=#1E90FF;display-name=fab0x;emotes=;flags=;id=e476375a-9685-475e-b515-c2aee821d998;login=fab0x;mod=0;msg-id=sub;msg-param-cumulative-months=1;msg-param-goal-contribution-type=SUB_POINTS;msg-param-goal-current-contributions=39024;msg-param-goal-target-contributions=77777;msg-param-goal-user-contributions=1;msg-param-months=0;msg-param-multimonth-duration=1;msg-param-multimonth-tenure=0;msg-param-should-share-streak=0;msg-param-sub-plan-name=Download\sA\sPirate;msg-param-sub-plan=Prime;msg-param-was-gifted=false;room-id=151368796;subscriber=1;system-msg=fab0x\ssubscribed\swith\sPrime.;tmi-sent-ts=1722278537616;user-id=29191648;user-type=;vip=0 :tmi.twitch.tv USERNOTICE #piratesoftware`
)

func main() {
	chat.Start()

	// timestamp := lipgloss.NewStyle().Faint(true).Render("[03:14]")
	// un := lipgloss.NewStyle().Foreground(lipgloss.Color("#9ACD32")).Render("Kosta")
	// msg := `To Reproduce Steps to reproduce the behavior:	Simply get the resize width and height, make sure it updated, center a string the most basic and classic way return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, fmt.Sprintf("%v-%v", m.width, m.height)). If it works on Linux and MacOs, perfect, now try on W10 on the Command Prompt. You should see the string going up.`
	// msg = wordwrap.String(msg, 60)
	// msgHeight := lipgloss.Height(msg)
	// var newT string = timestamp
	// for i := 1; i < msgHeight; i++ {
	// 	newT += "\n" + strings.Repeat(" ", lipgloss.Width(timestamp))
	// }
	// fmt.Println(lipgloss.JoinHorizontal(1, newT, " ", msg))
	//////////////////////////////////////////

	// fmt.Println(prime)
	// msgChan := make(chan interface{}, 100)
	// chat.ParseUSERNOTICE(prime, msgChan)
	// ws, err := chat.CreateWSClient()
	// if err != nil {
	// 	panic(err)
	// }
	// go ws.Connect(accessToken, name, "jasontheween", msgChan)
	// for {
	// 	select {
	// 	case m := <-msgChan:
	// 		b, _ := json.MarshalIndent(m, "", " ")
	// 		fmt.Println(string(b))
	// 	}
	// }
}

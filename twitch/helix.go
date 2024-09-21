package twitch

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/Kostaaa1/twitchdl/types"
)

func (c *Client) GetUserInfo(loginName string) (*types.UserData, error) {
	u := fmt.Sprintf("%s/users?login=%s", c.helixURL, loginName)
	req, err := c.NewGetRequest(u)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Client-Id", c.config.Creds.ClientID)
	req.Header.Set("Authorization", c.GetToken())
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type data struct {
		Data []types.UserData `json:"data"`
	}
	var user data
	if err := json.Unmarshal(b, &user); err != nil {
		return nil, err
	}
	if len(user.Data) == 0 {
		return nil, fmt.Errorf("the channel %s does not exist", loginName)
	}
	return &user.Data[0], nil
}

func (c *Client) GetChannelInfo(broadcasterID string) (*types.ChannelData, error) {
	u := fmt.Sprintf("%s/channels?broadcaster_id=%s", c.helixURL, broadcasterID)
	req, err := c.NewGetRequest(u)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Client-Id", c.config.Creds.ClientID)
	req.Header.Set("Authorization", c.GetToken())
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type data struct {
		Data []types.ChannelData `json:"data"`
	}
	var channel data
	if err := json.Unmarshal(b, &channel); err != nil {
		return nil, err
	}
	return &channel.Data[0], nil
}

func (c *Client) GetFollowedStreams(id string) (*types.Streams, error) {
	u := fmt.Sprintf("%s/streams/followed?user_id=%s", c.helixURL, id)
	req, err := c.NewGetRequest(u)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Client-Id", c.config.Creds.ClientID)
	req.Header.Set("Authorization", c.GetToken())
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var streams types.Streams
	if err := json.Unmarshal(b, &streams); err != nil {
		return nil, err
	}
	return &streams, nil
}

func (c *Client) GetStream(userId string) (*types.Streams, error) {
	u := fmt.Sprintf("%s/streams?user_id=%s", c.helixURL, userId)
	req, err := c.NewGetRequest(u)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Client-Id", c.config.Creds.ClientID)
	req.Header.Set("Authorization", c.GetToken())
	resp, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var streams types.Streams
	if err := json.Unmarshal(b, &streams); err != nil {
		return nil, err
	}
	return &streams, nil
}

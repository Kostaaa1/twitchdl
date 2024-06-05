package twitch

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	utils "github.com/Kostaaa1/twitch-clip-downloader/utils/file"
)

const (
	gqlURL      = "https://gql.twitch.tv/gql"
	gqlClientID = "kimne78kx3ncx6brgo4mv6wki5h1ko"
	usherURL    = "https://usher.ttvnw.net"
	helixURL    = ""
)

type Client struct {
	client      *http.Client
	clientID    string
	gqlURL      string
	helixURL    string
	usherURL    string
	gqlClientID string
}

type VideoType int

const (
	TypeClip VideoType = iota
	TypeVOD
)

func (c *Client) Name(vType VideoType, id string) (string, error) {
	var name string
	switch vType {
	case TypeClip:
		clip, err := c.ClipMetadata(id)
		if err != nil {
			fmt.Println("error", err)
			return "", err
		}
		name = fmt.Sprintf("%s - %s", clip.Broadcaster.DisplayName, clip.Title)
	case TypeVOD:
		clip, err := c.VideoMetadata(id)
		if err != nil {
			fmt.Println("error", err)
			return "", err
		}
		name = fmt.Sprintf("%s - %s", clip.Owner.Login, clip.Title)
	}

	return name, nil
}

func (c *Client) PathName(vType VideoType, id, output string) (string, error) {
	name, err := c.Name(vType, id)
	if err != nil {
		return "", err
	}
	name = utils.CreatePathname(output, name)
	return name, nil
}

func New(client *http.Client, clientID string) Client {
	return Client{
		client:      client,
		clientID:    clientID,
		gqlURL:      gqlURL,
		gqlClientID: gqlClientID,
		usherURL:    usherURL,
		helixURL:    helixURL,
	}
}

func (c *Client) do(req *http.Request, p interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		msg := fmt.Errorf("failed to perform request: %s", err)
		return msg
	}
	defer resp.Body.Close()
	if s := resp.StatusCode; s < 200 || s >= 300 {
		b, _ := io.ReadAll(resp.Body)
		fmt.Println("BODYYYY: ", string(b))
		return fmt.Errorf("unsupported status code: %v", s)
	}
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return fmt.Errorf("failed to decode response body: %s", err)
	}
	return nil
}

func (c *Client) ID(URL string) (string, VideoType, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse the URL: %s", err)
	}
	if !strings.Contains(u.Hostname(), "twitch.tv") {
		return "", 0, fmt.Errorf("the hostname of the URL does not contain twitch.tv")
	}
	if strings.Contains(u.Path, "/clip/") {
		_, id := path.Split(u.Path)
		return id, TypeClip, nil
	}
	if strings.Contains(u.Path, "/videos/") {
		_, id := path.Split(u.Path)
		return id, TypeVOD, nil
	}
	return "", 0, nil
}

func (c *Client) responsePreview(req *http.Request) {
	resp, err := c.client.Do(req)
	if err != nil {
		log.Fatalf("failed to get the response body: %s", err)
	}
	defer resp.Body.Close()
	byteResp, _ := io.ReadAll(resp.Body)
	var p map[string]interface{}
	if err := json.Unmarshal(byteResp, &p); err != nil {
		log.Fatalf("failed to unmarshal the body bytes: %s", err)
	}
	formatted, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		log.Fatalf("failed to marshal indent the JSON: %s", err)
	}
	fmt.Println("PREVIEW FORMATTED RESPONSE: ", string(formatted))
}

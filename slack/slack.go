package slack

import (
	"context"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/nlopes/slack"
)

// Client is a chat client
type Client struct {
	apiClient    *slack.Client
	rtm          *slack.RTM
	messageMatch func(string) bool
}

// ClientConfig client configuration used to setup the Slack client
type ClientConfig struct {
	Token string
	Debug bool
}

// New builds a new chat client
func New(conf ClientConfig) (*Client, error) {
	slackAPI := slack.New(conf.Token)
	slackAPI.SetDebug(conf.Debug)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := slackAPI.AuthTestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not connect to slack: %s", err)
	}

	rtm := slackAPI.NewRTM()
	go rtm.ManageConnection()

	return &Client{
		apiClient: slackAPI,
		rtm:       rtm,
		messageMatch: func(message string) bool {
			botUser := fmt.Sprintf("<@%s>", rtm.GetInfo().User.ID)
			return strings.HasPrefix(message, botUser)
		},
	}, nil
}

// ListenMessages listens to messages and sends the matching ones through the channel
func (c *Client) ListenMessages(ch chan Message) {
	log.Infof("Listening messages")

	for msg := range c.rtm.IncomingEvents {

		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if c.messageMatch(ev.Text) {
				log.Println("Received matching message", ev.Text)
				ch <- Message{
					Text:    ev.Text,
					Channel: ev.Channel,
					From:    ev.User,
				}
			}
		default:
			log.Debugf("Received Slack Event %#v\n", ev)
		}

	}
}

// Reply sends a message to a channel
func (c *Client) Reply(text, channel string) {
	msg := c.rtm.NewOutgoingMessage(text, channel)
	c.rtm.SendMessage(msg)
}

// ReplyIM sends a message to a user over an IM channel
func (c *Client) ReplyIM(text, user string) error {
	_, _, channel, err := c.apiClient.OpenIMChannel(user)
	if err != nil {
		return fmt.Errorf("could not open IM with %s: %s", user, err)
	}
	msg := c.rtm.NewOutgoingMessage(text, channel)
	c.rtm.SendMessage(msg)
	return nil
}

// Message a chat message
type Message struct {
	Text    string
	Channel string
	From    string
}

// GetText returns the message text
func (m Message) GetText() string {
	return m.Text
}

// GetUserFrom returns the user id formatted for using in a slack message
func (m Message) GetUserFrom() string {
	return formatSlackUser(m.From)
}

// GetChannel returns the channel from which the message was sent
func (m Message) GetChannel() string {
	return m.Channel
}

// GetUserFromID returns the raw user ID
func (m Message) GetUserFromID() string {
	return m.From
}

func formatSlackUser(userID string) string {
	return fmt.Sprintf("<@%s>", userID)
}

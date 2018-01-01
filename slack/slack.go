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
	messageMatch func(string) (bool, int)
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
		messageMatch: func(message string) (bool, int) {
			botUser := fmt.Sprintf("<@%s>", rtm.GetInfo().User.ID)
			return strings.HasPrefix(message, botUser), len(botUser)
		},
	}, nil
}

// ListenMessages listens to messages and sends the matching ones through the channel
func (c *Client) ListenMessages(ch chan Message) {
	log.Infof("Listening messages")

	for msg := range c.rtm.IncomingEvents {

		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if match, length := c.messageMatch(ev.Text); match {
				log.Infof("Received matching message", ev.Text)
				ch <- Message{
					Text:     strings.TrimSpace(ev.Text[length:]),
					Channel:  ev.Channel,
					ReplyTo:  ev.User,
					Username: ev.Username,
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
	Text     string
	Channel  string
	ReplyTo  string
	Username string
}

// GetText returns the message text
func (m Message) GetText() string {
	return m.Text
}

// GetReplyTo returns the user id formatted for using in a slack message
func (m Message) GetReplyTo() string {
	return fmt.Sprintf("<@%s>", m.ReplyTo)
}

// GetUsername returns the user friendly username
func (m Message) GetUsername() string {
	return m.Username
}

// GetChannel returns the channel from which the message was sent
func (m Message) GetChannel() string {
	return m.Channel
}

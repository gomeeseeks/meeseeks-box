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
	messageMatch func(*slack.MessageEvent) (bool, int)
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
		messageMatch: func(message *slack.MessageEvent) (bool, int) {
			botID := rtm.GetInfo().User.ID
			if message.User == botID {
				return false, 0 // It's myself talking
			}
			if strings.HasPrefix(message.Channel, "D") {
				// if the first letter of the channel is a D, it's an IM channel
				return true, 0
			}

			botUser := fmt.Sprintf("<@%s>", botID)
			return strings.HasPrefix(message.Text, botUser), len(botUser)
		},
	}, nil
}

// ListenMessages listens to messages and sends the matching ones through the channel
func (c *Client) ListenMessages(ch chan Message) {
	log.Infof("Listening messages")

	for msg := range c.rtm.IncomingEvents {

		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if match, length := c.messageMatch(ev); match {
				log.Infof("Received matching message", ev.Text)
				u, err := c.rtm.GetUserInfo(ev.User)
				if err != nil {
					log.Errorf("could not find user with id %s because %s, weeeird", ev.User, err)
				}
				ch <- Message{
					Text:     strings.TrimSpace(ev.Text[length:]),
					Channel:  ev.Channel,
					ReplyTo:  ev.User,
					Username: u.Name,
				}
			}
		default:
			log.Debugf("Received Slack Event %#v\n", ev)
		}

	}
}

// Reply replies to the user building a message with attachment
func (c *Client) Reply(content, color, channel string) error {
	params := slack.PostMessageParameters{
		AsUser: true,
		Attachments: []slack.Attachment{
			slack.Attachment{
				Text:       content,
				Color:      color,
				MarkdownIn: []string{"text"},
			},
		},
	}
	_, _, err := c.apiClient.PostMessage(channel, "", params)
	return err
}

// ReplyIM sends a message to a user over an IM channel
func (c *Client) ReplyIM(content, color, user string) error {
	_, _, channel, err := c.apiClient.OpenIMChannel(user)
	if err != nil {
		return fmt.Errorf("could not open IM with %s: %s", user, err)
	}
	return c.Reply(content, color, channel)
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

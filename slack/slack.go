package slack

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/nlopes/slack"
)

// Client is a chat client
type Client struct {
	apiClient    *slack.Client
	rtm          *slack.RTM
	messageMatch func(*slack.MessageEvent) *Message
}

// Connect builds a new chat client
func Connect(debug bool) (*Client, error) {
	token := os.Getenv("SLACK_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("could not connect to slack: SLACK_TOKEN env var is empty")
	}

	slackAPI := slack.New(token)
	slackAPI.SetDebug(debug)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := slackAPI.AuthTestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not connect to slack: %s", err)
	}

	rtm := slackAPI.NewRTM()
	go rtm.ManageConnection()

	mm := newMessageMatcher(rtm)
	return &Client{
		apiClient:    slackAPI,
		rtm:          rtm,
		messageMatch: mm.Matches,
	}, nil
}

type messageMatcher struct {
	botID         string
	prefixMatches []string
	rtm           *slack.RTM
}

func newMessageMatcher(rtm *slack.RTM) messageMatcher {
	return messageMatcher{
		rtm: rtm,
	}
}

func (m messageMatcher) Matches(message *slack.MessageEvent) *Message {
	if m.botID == "" {
		m.botID = m.rtm.GetInfo().User.ID
		m.prefixMatches = []string{fmt.Sprintf("<@%s>", m.botID)}
	}
	if text, ok := m.messageToMe(message); ok {
		return &Message{
			text:    text,
			channel: message.Channel,
			replyTo: message.User,
			isIM:    m.isIMChannel(message),
		}
	}
	return nil
}

func (m messageMatcher) isMyself(message *slack.MessageEvent) bool {
	return message.User == m.botID
}

func (m messageMatcher) isIMChannel(message *slack.MessageEvent) bool {
	return strings.HasPrefix(message.Channel, "D")
}

func (m messageMatcher) messageToMe(message *slack.MessageEvent) (string, bool) {
	if m.isMyself(message) {
		return "", false
	}
	if m.isIMChannel(message) {
		return message.Text, true
	}
	for _, match := range m.prefixMatches {
		if strings.HasPrefix(message.Text, match) {
			return strings.TrimSpace(message.Text[len(match):]), true
		}
	}
	return "", false
}

// ListenMessages listens to messages and sends the matching ones through the channel
func (c *Client) ListenMessages(ch chan<- Message) {
	log.Infof("Listening messages")

	for msg := range c.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			message := c.messageMatch(ev)
			if message == nil {
				continue
			}

			log.Debugf("Received matching message %#v", ev.Text)
			u, err := c.rtm.GetUserInfo(ev.User)
			if err != nil {
				log.Errorf("could not find user with id %s because %s, weeeird", ev.User, err)
				continue
			}
			message.username = u.Name

			ch <- *message

		default:
			log.Debugf("Ignored Slack Event %#v", ev)
		}
	}
	log.Infof("Stopped listening to messages")
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
	text     string
	channel  string
	replyTo  string
	username string
	isIM     bool
}

// GetText returns the message text
func (m Message) GetText() string {
	return m.text
}

// GetReplyTo returns the user id formatted for using in a slack message
func (m Message) GetReplyTo() string {
	return fmt.Sprintf("<@%s>", m.replyTo)
}

// GetUsername returns the user friendly username
func (m Message) GetUsername() string {
	return m.username
}

// GetChannel returns the channel from which the message was sent
func (m Message) GetChannel() string {
	return m.channel
}

// IsIM returns if the message is an IM message
func (m Message) IsIM() bool {
	return m.isIM
}

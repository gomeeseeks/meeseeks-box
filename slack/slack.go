package slack

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pcarranza/meeseeks-box/meeseeks/message"
	"github.com/sirupsen/logrus"

	"github.com/nlopes/slack"
)

var errIgnoredMessage = fmt.Errorf("Ignore this message")

// Client is a chat client
type Client struct {
	apiClient *slack.Client
	// TODO: remove the rtm as it should only be inside the message matcher.
	// It should simply be inside there and it should pop messages matched out
	// through a channel
	rtm     *slack.RTM
	matcher messageMatcher
}

// GetUsername implements the messenger.MessengerClient interface
func (c Client) GetUsername(userID string) string {
	return c.matcher.getUser(userID)
}

// GetUserLink implements the messenger.MessengerClient interface
func (c Client) GetUserLink(userID string) string {
	return fmt.Sprintf("<@%s>", userID)
}

// GetChannel implements the messenger.MessengerClient interface
func (c Client) GetChannel(channelID string) string {
	return c.matcher.getChannel(channelID)
}

// GetChannelLink implements the messenger.MessengerClient interface
func (c Client) GetChannelLink(channelID string) string {
	return fmt.Sprintf("<#%s|%s>", channelID, c.matcher.getChannel(channelID))
}

// IsIM implements the messenger.MessengerClient interface
func (c Client) IsIM(channelID string) bool {
	return c.matcher.isIMChannel(channelID)
}

// Connect builds a new chat client
func Connect(debug bool, token string) (*Client, error) {
	if token == "" {
		return nil, fmt.Errorf("could not connect to slack: SLACK_TOKEN env var is empty")
	}

	slackClient := slack.New(token)
	slackClient.SetDebug(debug)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := slackClient.AuthTestContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not connect to slack: %s", err)
	}

	rtm := slackClient.NewRTM()
	go rtm.ManageConnection()

	return &Client{
		apiClient: slackClient,
		rtm:       rtm,
		matcher:   newMessageMatcher(rtm),
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

// GetUser finds the username given a userID
func (m *messageMatcher) getUser(userID string) string {
	u, err := m.rtm.GetUserInfo(userID)
	if err != nil {
		logrus.Errorf("could not find user with id %s because %s, weeeird", userID, err)
		return "unknown-user"
	}
	return u.Name
}

func (m *messageMatcher) isIMChannel(channel string) bool {
	return strings.HasPrefix(channel, "D")
}

// GetChannel returns a channel name given an ID
func (m *messageMatcher) getChannel(channelID string) string {
	if m.isIMChannel(channelID) {
		return "IM"
	}

	ch, err := m.rtm.GetChannelInfo(channelID)
	if err != nil {
		logrus.Errorf("could not find channel with id %s: %s", channelID, err)
		return "unknown-channel"
	}
	return ch.Name
}

// Init has to be delayed until the point in which the RTM is actually working.
// The simples way to do this lazily is to do it when the message listening starts
func (m *messageMatcher) init() {
	if m.botID == "" {
		m.botID = m.rtm.GetInfo().User.ID
		m.prefixMatches = []string{fmt.Sprintf("<@%s>", m.botID)}
	}
}

func (m *messageMatcher) Matches(message *slack.MessageEvent) (Message, error) {
	m.init()

	if text, ok := m.shouldCare(message); ok {
		username := m.getUser(message.User)
		channel := m.getChannel(message.Channel)
		isIM := m.isIMChannel(message.Channel)

		return Message{
			text:      text,
			userID:    message.User,
			channelID: message.Channel,
			username:  username,
			channel:   channel,
			isIM:      isIM,
		}, nil
	}
	return Message{}, errIgnoredMessage
}

func (m *messageMatcher) isMyself(message *slack.MessageEvent) bool {
	return message.User == m.botID
}

func (m *messageMatcher) shouldCare(message *slack.MessageEvent) (string, bool) {
	if m.isMyself(message) {
		logrus.Debug("It's myself, ignoring message")
		return "", false
	}
	if m.isIMChannel(message.Channel) {
		logrus.Debugf("Channel %s is IM channel, responding...", message.Channel)
		return message.Text, true
	}
	for _, match := range m.prefixMatches {
		if strings.HasPrefix(message.Text, match) {
			logrus.Debugf("Message '%s' matches prefix, responding...", message.Text)
			return strings.TrimSpace(message.Text[len(match):]), true
		}
	}
	return "", false
}

// ListenMessages listens to messages and sends the matching ones through the channel
func (c *Client) ListenMessages(ch chan<- message.Message) {
	logrus.Infof("Listening Slack RTM Messages")

	for msg := range c.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			message, err := c.matcher.Matches(ev)
			if err != nil {
				continue
			}

			logrus.Debugf("Sending Slack message %#v to messages channel", message)
			ch <- message

		default:
			logrus.Debugf("Ignored Slack Event %#v", ev)
		}
	}
	logrus.Infof("Stopped listening to messages")
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
	logrus.Debugf("Replying in Slack %s with %#v", channel, params)
	_, _, err := c.apiClient.PostMessage(channel, "", params)
	return err
}

// ReplyIM sends a message to a user over an IM channel
func (c *Client) ReplyIM(content, color, user string) error {
	_, _, channel, err := c.apiClient.OpenIMChannel(user)
	if err != nil {
		return fmt.Errorf("could not open IM with %s: %s", user, err)
	}
	logrus.Debugf("Replying in Slack IM with '%s' and color %s", content, color)
	return c.Reply(content, color, channel)
}

// Message a chat message
type Message struct {
	text      string
	channel   string
	channelID string
	username  string
	userID    string
	isIM      bool
}

// GetText returns the message text
func (m Message) GetText() string {
	return m.text
}

// GetUserID returns the user ID
func (m Message) GetUserID() string {
	return m.userID
}

// GetUserLink returns the user id formatted for using in a slack message
func (m Message) GetUserLink() string {
	return fmt.Sprintf("<@%s>", m.userID)
}

// GetUsername returns the user friendly username
func (m Message) GetUsername() string {
	return m.username
}

// GetChannelID returns the channel id from the which the message was sent
func (m Message) GetChannelID() string {
	return m.channelID
}

// GetChannel returns the channel from which the message was sent
func (m Message) GetChannel() string {
	return m.channel
}

// GetChannelLink returns the channel that slack will turn into a link
func (m Message) GetChannelLink() string {
	return fmt.Sprintf("<#%s|%s>", m.channelID, m.channel)
}

// IsIM returns if the message is an IM message
func (m Message) IsIM() bool {
	return m.isIM
}

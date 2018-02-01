package messenger

import (
	"fmt"

	"github.com/pcarranza/meeseeks-box/meeseeks/message"
	"github.com/pcarranza/meeseeks-box/slack"
)

// Messenger handles multiple message sources
type Messenger struct {
	*slack.Client
	messagesCh chan message.Message
}

type MessengerOpts struct {
	Debug      bool
	SlackToken string
}

func Listen(opts MessengerOpts) (*Messenger, error) {
	client, err := slack.Connect(opts.Debug, opts.SlackToken)
	if err != nil {
		return nil, fmt.Errorf("could not connect to slack: %s", err)
	}

	slackMessagesCh := make(chan message.Message)
	go client.ListenMessages(slackMessagesCh)

	return &Messenger{
		Client:     client,
		messagesCh: slackMessagesCh,
	}, nil
}

func (m *Messenger) MessagesCh() chan message.Message {
	return m.messagesCh
}

func (m *Messenger) Shutdown() {
	close(m.messagesCh)
}

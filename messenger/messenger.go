package messenger

import (
	"github.com/pcarranza/meeseeks-box/meeseeks/message"
	"github.com/sirupsen/logrus"
)

// Listener provides the necessary interface to start listening messages in a channel.
type Listener interface {
	ListenMessages(chan<- message.Message)
}

// Messenger handles multiple message sources
type Messenger struct {
	messagesCh chan message.Message
}

// Listen starts a routine to listen for messages on the provided client
func Listen(listeners ...Listener) (*Messenger, error) {
	messagesCh := make(chan message.Message)

	for _, listener := range listeners {
		go listener.ListenMessages(messagesCh)
	}

	return &Messenger{
		messagesCh: messagesCh,
	}, nil
}

// MessagesCh returns the channel in which to listen for messages
func (m *Messenger) MessagesCh() <-chan message.Message {
	return m.messagesCh
}

// Shutdown takes down the system
func (m *Messenger) Shutdown() {
	logrus.Infof("Shutting down messenger messages channel")
	close(m.messagesCh)
}

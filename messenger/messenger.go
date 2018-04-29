package messenger

// import (
// 	"github.com/gomeeseeks/meeseeks-box/meeseeks"
// 	"github.com/sirupsen/logrus"
// )

// // Messenger handles multiple message sources
// type Messenger struct {
// 	messagesCh chan meeseeks.Request
// }

// // Listen starts a routine to listen for messages on the provided client
// func Listen(listeners ...Listener) (*Messenger, error) {
// 	messagesCh := make(chan meeseeks.Request)

// 	for _, listener := range listeners {
// 		go listener.ListenMessages(messagesCh)
// 	}

// 	return &Messenger{
// 		messagesCh: messagesCh,
// 	}, nil
// }

// // Shutdown takes down the system
// func (m *Messenger) Shutdown() {
// 	logrus.Infof("Shutting down messenger messages channel")
// 	close(m.messagesCh)
// }

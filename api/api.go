package api

import (
	"net/http"

	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"

	"github.com/pcarranza/meeseeks-box/meeseeks/message"
	"github.com/pcarranza/meeseeks-box/tokens"
)

// MetadataClient is a helper client used to augment the metadata of the user
// and channel extracted from the registered token.
type MetadataClient interface {
	GetUser(string) string
	GetChannel(string) string
	IsIM(string) bool
}

// Server is used to provide API access
type Server struct {
	metadata  MetadataClient
	messageCh chan message.Message
}

// NewServer returns a new API Server that will use the provided metadata client
func NewServer(client MetadataClient) Server {
	return Server{
		metadata:  client,
		messageCh: make(chan message.Message),
	}
}

// Listen starts listening
func (s Server) Listen(path, address string) {
	http.HandleFunc(path, s.HandlePostToken)
	go http.ListenAndServe(address, nil)
}

// HandlePostToken handles a request
func (s Server) HandlePostToken(w http.ResponseWriter, r *http.Request) {
	tokenID := r.Header.Get("TOKEN")
	if tokenID == "" {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}
	logrus.Debugf("received token %s through API", tokenID) // Add requester info

	token, err := tokens.Get(tokenID)
	if err != nil {
		logrus.Debugf("Token %s is unknown", token) // Add requester info
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	log.Infof("token is %#v", token)

	logrus.Infof("received valid token %s from", token)
	s.messageCh <- apiMessage{
		metadata: s.metadata,
		token:    token,
	}

	w.WriteHeader(http.StatusAccepted)
}

// ListenMessages listens to messages and sends the matching ones through the channel
func (s Server) ListenMessages(ch chan<- message.Message) {
	for m := range s.messageCh {
		ch <- m
	}
}

// Message a chat message
type apiMessage struct {
	token    tokens.Token
	metadata MetadataClient
}

// GetText returns the message text
func (m apiMessage) GetText() string {
	return m.token.Text
}

// GetUsernameID returns the user id formatted for using in a slack message
func (m apiMessage) GetUsernameID() string {
	return m.token.User
}

// GetUsername returns the user friendly username
func (m apiMessage) GetUsername() string {
	return m.metadata.GetUser(m.token.User)
}

// GetChannelID returns the channel id from the which the message was sent
func (m apiMessage) GetChannelID() string {
	return m.token.Channel
}

// GetChannel returns the channel from which the message was sent
func (m apiMessage) GetChannel() string {
	return m.metadata.GetChannel(m.token.Channel)
}

// GetChannelLink returns the channel that slack will turn into a link
func (m apiMessage) GetChannelLink() string {
	return m.metadata.GetChannel(m.token.Channel)
}

// IsIM returns if the message is an IM message
func (m apiMessage) IsIM() bool {
	return m.metadata.IsIM(m.token.Channel)
}

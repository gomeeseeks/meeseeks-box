package api

import (
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/pcarranza/meeseeks-box/meeseeks/message"
	"github.com/pcarranza/meeseeks-box/tokens"
)

// MetadataClient is a helper client used to augment the metadata of the user
// and channel extracted from the registered token.
type MetadataClient interface {
	GetUsername(string) string
	GetUserLink(string) string
	GetChannel(string) string
	GetChannelLink(string) string
	IsIM(string) bool
}

// Listener implements the message Listener API and is used to send the messaged
// received from the API to the messaging pipeline
type Listener struct {
	metadata  MetadataClient
	messageCh chan message.Message
}

func (l Listener) sendMessage(token tokens.Token, message string) {
	l.messageCh <- apiMessage{
		metadata:    l.metadata,
		token:       token,
		formMessage: message,
	}
}

// ListenMessages listens to messages and sends the matching ones through the channel
func (l Listener) ListenMessages(ch chan<- message.Message) {
	for m := range l.messageCh {
		ch <- m
	}
}

// Shutdown closes the internal messages channel
func (l Listener) Shutdown() {
	logrus.Infof("Shutting down API messages channel")
	close(l.messageCh)
}

// NewListener returns a new message listener unsing the provided metadata client
func NewListener(client MetadataClient) Listener {
	return Listener{
		metadata:  client,
		messageCh: make(chan message.Message),
	}
}

// Server is used to provide API access
type Server struct {
	listener   Listener
	httpServer http.Server
}

// NewServer returns a new API Server that will use the provided metadata client
func NewServer(client MetadataClient, address string) Server {
	return Server{
		listener: NewListener(client),
		httpServer: http.Server{
			Addr: address,
		},
	}
}

// ListenAndServe starts listening on the provided address, then serving http requests
func (s *Server) ListenAndServe(path string) error {
	http.HandleFunc(path, s.HandlePostToken)
	return s.httpServer.ListenAndServe()
}

// GetListener returns the internal messages listener to register with the chat pipeline
func (s *Server) GetListener() Listener {
	return s.listener
}

// Shutdown shuts down the http server gracefully
func (s *Server) Shutdown() error {
	defer s.listener.Shutdown()
	logrus.Infof("Shutting down API server")

	return s.httpServer.Shutdown(nil)
}

// HandlePostToken handles a request
func (s *Server) HandlePostToken(w http.ResponseWriter, r *http.Request) {
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

	logrus.Infof("received valid token %s from", token)
	s.listener.sendMessage(token, r.FormValue("message"))

	w.WriteHeader(http.StatusAccepted)
}

// Message a chat message
type apiMessage struct {
	token       tokens.Token
	metadata    MetadataClient
	formMessage string
}

// GetText returns the message text
func (m apiMessage) GetText() string {
	text := m.token.Text
	if m.formMessage != "" {
		text = strings.Join([]string{text, m.formMessage}, " ")
	}
	return text
}

// GetUsernameID returns the user id formatted for using in a slack message
func (m apiMessage) GetUserID() string {
	return m.token.UserID
}

// GetUsername returns the user friendly username
func (m apiMessage) GetUsername() string {
	return m.metadata.GetUsername(m.token.UserID)
}

// GetUserLink
func (m apiMessage) GetUserLink() string {
	return m.metadata.GetUserLink(m.token.UserID)
}

// GetChannelID returns the channel id from the which the message was sent
func (m apiMessage) GetChannelID() string {
	return m.token.ChannelID
}

// GetChannel returns the channel from which the message was sent
func (m apiMessage) GetChannel() string {
	return m.metadata.GetChannel(m.token.ChannelID)
}

// GetChannelLink returns the channel that slack will turn into a link
func (m apiMessage) GetChannelLink() string {
	return m.metadata.GetChannelLink(m.token.ChannelID)
}

// IsIM returns if the message is an IM message
func (m apiMessage) IsIM() bool {
	return m.metadata.IsIM(m.token.ChannelID)
}

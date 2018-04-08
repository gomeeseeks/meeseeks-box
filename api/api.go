package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/tokens"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetadataClient is a helper client used to augment the metadata of the user
// and channel extracted from the registered token.
type MetadataClient interface {
	ParseChannelLink(string) (string, error)
	ParseUserLink(string) (string, error)
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
	messageCh chan meeseeks.Message
}

func (l Listener) sendMessage(token meeseeks.APIToken, message string) {
	channelID, err := l.metadata.ParseChannelLink(token.ChannelLink)
	if err != nil {
		logrus.Errorf("Failed to parse channel link %s: %s. Dropping message!", token.ChannelLink, err)
		return
		// TODO: this error should go to the administration channel
	}

	userID, err := l.metadata.ParseUserLink(token.UserLink)
	if err != nil {
		logrus.Errorf("Failed to parse user link %s: %s. Dropping message!", token.UserLink, err)
		return
		// TODO: this error should go to the administration channel
	}

	m := apiMessage{
		channelID:      channelID,
		userID:         userID,
		text:           token.Text,
		metadata:       l.metadata,
		messagePayload: message,
	}
	logrus.Debugf("Sending API message %#v to messages channel", m)
	l.messageCh <- m
}

// ListenMessages listens to messages and sends the matching ones through the channel
func (l Listener) ListenMessages(ch chan<- meeseeks.Message) {
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
		messageCh: make(chan meeseeks.Message),
	}
}

// Server is used to provide API access
type Server struct {
	listener   Listener
	httpServer http.Server
}

// NewServer returns a new API Server that will use the provided metadata client
func NewServer(client MetadataClient, metricsPath, apiPath, address string) *Server {
	s := Server{
		listener: NewListener(client),
		httpServer: http.Server{
			Addr: address,
		},
	}
	http.HandleFunc(apiPath, s.HandlePostToken)
	http.Handle(metricsPath, promhttp.Handler())
	return &s
}

// ListenAndServe starts listening on the provided address, then serving http requests
func (s *Server) ListenAndServe() error {
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

	return s.httpServer.Shutdown(context.TODO())
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

	s.listener.sendMessage(token, r.FormValue("message"))

	w.WriteHeader(http.StatusAccepted)
}

// Message a chat message
type apiMessage struct {
	userID         string
	channelID      string
	text           string
	messagePayload string
	metadata       MetadataClient
}

// GetText returns the message text
func (m apiMessage) GetText() string {
	text := m.text
	if m.messagePayload != "" {
		text = strings.Join([]string{text, m.messagePayload}, " ")
	}
	return text
}

// GetUsernameID returns the user id formatted for using in a slack message
func (m apiMessage) GetUserID() string {
	return m.userID
}

// GetUsername returns the user friendly username
func (m apiMessage) GetUsername() string {
	return m.metadata.GetUsername(m.userID)
}

// GetUserLink
func (m apiMessage) GetUserLink() string {
	return m.metadata.GetUserLink(m.userID)
}

// GetChannelID returns the channel id from the which the message was sent
func (m apiMessage) GetChannelID() string {
	return m.channelID
}

// GetChannel returns the channel from which the message was sent
func (m apiMessage) GetChannel() string {
	return m.metadata.GetChannel(m.channelID)
}

// GetChannelLink returns the channel that slack will turn into a link
func (m apiMessage) GetChannelLink() string {
	return m.metadata.GetChannelLink(m.channelID)
}

// IsIM returns if the message is an IM message
func (m apiMessage) IsIM() bool {
	return m.metadata.IsIM(m.channelID)
}

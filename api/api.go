package api

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/tokens"
)

// Enricher is a helper client used to augment the metadata of the user
// and channel extracted from the registered token.
type Enricher interface {
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
	enricher   Enricher
	requestsCh chan meeseeks.Request
	shutdown   chan bool
}

func (l Listener) sendMessage(token meeseeks.APIToken, message string) error {
	channelID, err := l.enricher.ParseChannelLink(token.ChannelLink)
	if err != nil {
		logrus.Errorf("Failed to parse channel link %s: %s. Dropping message!", token.ChannelLink, err)
		// TODO: this error should go to the administration channel
		return err
	}

	userID, err := l.enricher.ParseUserLink(token.UserLink)
	if err != nil {
		logrus.Errorf("Failed to parse user link %s: %s. Dropping message!", token.UserLink, err)
		// TODO: this error should go to the administration channel
		return err
	}

	l.requestsCh <- meeseeks.Request{
		UserID:      userID,
		Username:    l.enricher.GetUsername(userID),
		UserLink:    l.enricher.GetUserLink(userID),
		ChannelID:   channelID,
		Channel:     l.enricher.GetChannel(channelID),
		ChannelLink: l.enricher.GetChannelLink(channelID),
		IsIM:        l.enricher.IsIM(channelID),
	}
	return nil
}

// Listen starts a listening on
func (l Listener) Listen(ch chan<- meeseeks.Request) {
	shutdown := false
	for !shutdown {
		select {
		case r := <-l.requestsCh:
			ch <- r
		case shutdown = <-l.shutdown:
			break
		}
	}
}

// Shutdown closes the internal messages channel
func (l Listener) Shutdown() {
	logrus.Infof("Shutting down API messages channel")
	l.shutdown <- true
	close(l.requestsCh)
}

// Server is used to provide API access
type Server struct {
	listener   Listener
	httpServer http.Server
}

// NewServer returns a new API Server that will use the provided metadata client
func NewServer(enricher Enricher, path, address string) *Server {
	s := Server{
		listener: Listener{
			enricher,
			make(chan meeseeks.Request),
			make(chan bool),
		},
		httpServer: http.Server{
			Addr: address,
		},
	}
	http.HandleFunc(path, s.HandlePostToken)
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

	if err := s.listener.sendMessage(token, r.FormValue("message")); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

	w.WriteHeader(http.StatusAccepted)
}

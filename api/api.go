package api

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/parser"
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

// Server is used to provide API access
type Server struct {
	httpServer http.Server

	enricher   Enricher
	requestsCh chan meeseeks.Request
	shutdown   chan bool
}

// NewServer returns a new API Server that will use the provided metadata client
func NewServer(enricher Enricher, path, address string) *Server {
	s := Server{
		http.Server{
			Addr: address,
		},
		enricher,
		make(chan meeseeks.Request),
		make(chan bool),
	}
	http.HandleFunc(path, s.HandlePostToken)
	return &s
}

// ListenAndServe starts listening on the provided address, then serving http requests
func (s *Server) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) sendMessage(token meeseeks.APIToken, message string) error {
	channelID, err := s.enricher.ParseChannelLink(token.ChannelLink)
	if err != nil {
		logrus.Errorf("Failed to parse channel link %s: %s. Dropping message!", token.ChannelLink, err)
		// TODO: this error should go to the administration channel
		return err
	}

	userID, err := s.enricher.ParseUserLink(token.UserLink)
	if err != nil {
		logrus.Errorf("Failed to parse user link %s: %s. Dropping message!", token.UserLink, err)
		// TODO: this error should go to the administration channel
		return err
	}

	args, err := parser.Parse(token.Text + " " + message)
	if err != nil {
		return err
	}

	s.requestsCh <- meeseeks.Request{
		Command:     args[0],
		Args:        args[1:],
		UserID:      userID,
		Username:    s.enricher.GetUsername(userID),
		UserLink:    s.enricher.GetUserLink(userID),
		ChannelID:   channelID,
		Channel:     s.enricher.GetChannel(channelID),
		ChannelLink: s.enricher.GetChannelLink(channelID),
		IsIM:        s.enricher.IsIM(channelID),
	}
	return nil
}

// Listen starts a listening on
func (s *Server) Listen(ch chan<- meeseeks.Request) {
	shutdown := false
	for !shutdown {
		select {
		case r := <-s.requestsCh:
			ch <- r
		case shutdown = <-s.shutdown:
			break
		}
	}
}

// Shutdown shuts down the http server gracefully
func (s *Server) Shutdown() error {
	logrus.Infof("Shutting down API messages channel")
	s.shutdown <- true
	close(s.requestsCh)

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

	if err := s.sendMessage(token, r.FormValue("message")); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

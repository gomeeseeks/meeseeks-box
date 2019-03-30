package api

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks"
	"gitlab.com/yakshaving.art/meeseeks-box/persistence"
	"gitlab.com/yakshaving.art/meeseeks-box/text/parser"
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

// Service provides a service suitable to manage command requests through the API:w
type Service struct {
	enricher   Enricher
	requestsCh chan meeseeks.Request
	shutdown   chan bool
}

// New returns a new API service instance
func New(enricher Enricher, path string) *Service {
	s := &Service{
		enricher,
		make(chan meeseeks.Request),
		make(chan bool),
	}
	http.HandleFunc(path, s.HandlePostToken)
	return s
}

func (s *Service) sendMessage(token meeseeks.APIToken, message string) error {
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

// Listen starts listen on the passed in channel
func (s *Service) Listen(ch chan<- meeseeks.Request) {
	shutdown := false
	for !shutdown {
		select {
		case r := <-s.requestsCh:
			ch <- r
		case shutdown = <-s.shutdown:
			// nothing to do here
		}
	}
}

// Shutdown shuts down the http server gracefully
func (s *Service) Shutdown() error {
	logrus.Infof("Shutting down API Service")
	s.shutdown <- true
	close(s.requestsCh)

	return nil
}

// HandlePostToken implements the http handle request function interface
func (s *Service) HandlePostToken(w http.ResponseWriter, r *http.Request) {
	tokenID := r.Header.Get("TOKEN")
	if tokenID == "" {
		http.Error(w, "no token", http.StatusBadRequest)
		return
	}
	logrus.Debugf("received token %s through API", tokenID) // Add requester info

	token, err := persistence.APITokens().Get(tokenID)
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

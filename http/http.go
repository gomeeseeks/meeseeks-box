package http

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Service is a simple interface used to manage a service lifecycle
type Service interface {
	Shutdown() error
}

// Server manages all the services lifecycle that work on http protocol
type Server struct {
	server   http.Server
	services []Service
}

// New returns a new API Server that will use the provided metadata client
func New(address string) *Server {
	s := Server{
		server: http.Server{
			Addr: address,
		},
		services: make([]Service, 0),
	}
	return &s
}

// ListenAndServe starts listening on http
func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Register starts and appends a service in the server
func (s *Server) Register(service Service) {
	s.services = append(s.services, service)
}

// Shutdown handles shutting down all the registered services
func (s *Server) Shutdown() error {
	for _, service := range s.services {
		if err := service.Shutdown(); err != nil {
			return err
		}
	}
	logrus.Infof("Shutting down HTTP server")
	s.server.Shutdown(context.Background())
	return nil
}

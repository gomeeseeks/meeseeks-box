package server

import (
	"fmt"
	"net"

	"github.com/gomeeseeks/meeseeks-box/remote/api"
	"google.golang.org/grpc"
)

// RemoteServer is a remote server, duh
type RemoteServer struct {
	Address string
	server  *grpc.Server
}

// New creates a new RemoteServer with an address
func New(addr string) *RemoteServer {

	s := grpc.NewServer()
	api.RegisterLogWriterServer(s, logWriterServer{})
	api.RegisterCommandPipelineServer(s, newCommandPipelineServer())

	return &RemoteServer{
		Address: addr,
		server:  s,
	}
}

// Listen starts the listening of a remote server
func (s RemoteServer) Listen() error {
	address, err := net.Listen("tcp", s.Address)
	if err != nil {
		return fmt.Errorf("could parse address %s: %s", s.Address, err)
	}

	if err := s.server.Serve(address); err != nil {
		return fmt.Errorf("failed to start listening on address %s: %s", s.Address, err)
	}
	return nil
}

// Shutdown stops listening for requests
func (s RemoteServer) Shutdown() {
	s.server.Stop()
}

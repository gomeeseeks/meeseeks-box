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

func New(addr string) RemoteServer {
	return RemoteServer{
		Address: addr,
		server:  grpc.NewServer(),
	}
}

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

// Register registers the different servers in the grpc server
func (s *RemoteServer) Register() {
	api.RegisterLogWriterServer(s.server, logWriterServer{})
	api.RegisterCommandPipelineServer(s.server, newCommandPipelineServer())
}

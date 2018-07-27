package server

import (
	"context"
	"errors"

	"github.com/gomeeseeks/meeseeks-box/persistence"
	"github.com/gomeeseeks/meeseeks-box/remote/api"
	"google.golang.org/grpc"
)

// RemoteServer is a remote server, duh
type RemoteServer struct {
	Address string
	server  *grpc.Server
}

// Register registers the different servers in the grpc server
func (s *RemoteServer) Register() {
	api.RegisterLogWriterServer(s.server, logWriterServer{})
	api.RegisterCommandPipelineServer(s.server, CommandPipelineServer{})
}

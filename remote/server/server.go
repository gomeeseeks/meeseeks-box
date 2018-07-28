package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gomeeseeks/meeseeks-box/remote/api"
	"google.golang.org/grpc"
)

// RemoteServer is a remote server, duh
type RemoteServer struct {
	server *grpc.Server
}

// New creates a new RemoteServer with an address
func New() *RemoteServer {
	s := grpc.NewServer()
	api.RegisterLogWriterServer(s, logWriterServer{})
	api.RegisterCommandPipelineServer(s, newCommandPipelineServer())

	return &RemoteServer{
		server: s,
	}
}

// Register registers the grpc server in an http path
func (s RemoteServer) Register(handlerPath string) {
	http.HandleFunc(handlerPath, s.server.ServeHTTP)
}

// Listen starts the listening of a remote server
func (s RemoteServer) Listen(addr string) error {
	address, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("could parse address %s: %s", addr, err)
	}

	if err := s.server.Serve(address); err != nil {
		return fmt.Errorf("failed to start listening on address %s: %s", addr, err)
	}
	return nil
}

// Shutdown stops listening for requests
func (s RemoteServer) Shutdown() {
	s.server.GracefulStop()
}

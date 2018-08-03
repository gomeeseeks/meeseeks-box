package server

import (
	"fmt"
	"net"

	"github.com/gomeeseeks/meeseeks-box/remote/api"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// TLS Security Model
const SecurityModeTLS = "tls"

// RemoteServer is a remote server, duh
type RemoteServer struct {
	server *grpc.Server
	config Config
}

// Config represents the grpc server configuration
type Config struct {
	CertPath     string
	KeyPath      string
	SecurityMode string
}

// New creates a new RemoteServer with an address
func New(c Config) (*RemoteServer, error) {

	options := []grpc.ServerOption{
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	}

	switch c.SecurityMode {
	case SecurityModeTLS:
		creds, err := credentials.NewServerTLSFromFile(c.CertPath, c.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("could not configure tls credentials: %s", err)
		}
		options = append(options, grpc.Creds(creds))

	default:
		logrus.Warnf("starting server in insecure mode (without encryption)")
	}

	s := grpc.NewServer(options...)

	api.RegisterLogWriterServer(s, logWriterServer{})
	api.RegisterCommandPipelineServer(s, newCommandPipelineServer())

	grpc_prometheus.Register(s)

	return &RemoteServer{
		server: s,
		config: c,
	}, nil
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
	logrus.Debugf("gracefully stopping grpc server")
	s.server.Stop()
	logrus.Debugf("grpc server stopped")
}

package main

import (
	"flag"
	"fmt"
	"net"

	pb "github.com/gomeeseeks/meeseeks-box/remote/api"
	"github.com/gomeeseeks/meeseeks-box/remote/server"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	port := flag.Int("port", 9996, "server port")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logrus.Fatalf("failed to start listening on port %d", *port)
	}

	s := grpc.NewServer()
	pb.RegisterCommandLoggerServer(s, server.CommandLoggerServer{})
	pb.RegisterCommandPipelineServer(s, server.CommandPipelineServer{})

	logrus.Infof("listening on port %d", *port)
	if err := s.Serve(lis); err != nil {
		logrus.Fatal(err)
	}
}

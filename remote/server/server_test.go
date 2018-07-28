package server_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/remote/server"
	"google.golang.org/grpc"
)

func TestAgentCanConnect(t *testing.T) {
	s := server.New()
	defer s.Shutdown()

	go func() {
		mocks.Must(t, "Failed to start server", s.Listen(":9699"))
	}()

	time.Sleep(1 * time.Millisecond) // Allow the server to start

	client, err := grpc.Dial("localhost:9699", grpc.WithInsecure())
	mocks.Must(t, "failed to execute HEAD command: %s", err)
	defer client.Close()

	fmt.Println(client.GetState())
}

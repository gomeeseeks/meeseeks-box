package provider

import (
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/persistence/logs/local"
	"github.com/gomeeseeks/meeseeks-box/persistence/logs/remote"
)

// Logger Constants
const (
	LocalLogger  = "local"
	RemoteLogger = "remote"
)

// New creates a new LoggerProvider matching the provided configuration
func New(logger string) meeseeks.LoggerProvider {
	if logger == RemoteLogger {
		return remote.New() // This will need to change once we have some form of grpc remote configuration
	}
	return local.New()
}

package commands

import (
	"fmt"
	"time"
)

var (
	ErrCommandNotFound    = fmt.Errorf("Can't find command")
	ErrUnknownCommandType = fmt.Errorf("Unknown command type")
)

// Defaults for commands
const (
	DefaultCommandTimeout = 60 * time.Second
)

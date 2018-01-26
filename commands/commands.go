package commands

import (
	"fmt"
	"time"

	"github.com/pcarranza/meeseeks-box/command"
)

// ErrCommandNotFound is returned when a command cannot be found
var ErrCommandNotFound = fmt.Errorf("Can't find command")

// Defaults for commands
const (
	DefaultCommandTimeout = 60 * time.Second
)

var commands map[string]command.Command

func init() {
	commands = make(map[string]command.Command)
}

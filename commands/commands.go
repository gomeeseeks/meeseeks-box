package commands

import (
	"fmt"

	"github.com/pcarranza/meeseeks-box/command"
)

// ErrCommandNotFound is returned when a command cannot be found
var ErrCommandNotFound = fmt.Errorf("Can't find command")

var commands map[string]command.Command

func init() {
	commands = make(map[string]command.Command)
}

package command

import (
	"time"

	"context"

	"github.com/pcarranza/meeseeks-box/jobs"
)

// Defaults for commands
const (
	DefaultCommandTimeout = 60 * time.Second
)

// Command is the base interface for any command
type Command interface {
	Execute(context.Context, jobs.Job) (string, error)
	Cmd() string
	HasHandshake() bool
	Templates() map[string]string
	AuthStrategy() string
	AllowedGroups() []string
	Args() []string
	Timeout() time.Duration
	Help() string
	Record() bool
}

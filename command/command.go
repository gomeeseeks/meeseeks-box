package command

import (
	"time"

	"gitlab.com/mr-meeseeks/meeseeks-box/jobs"
)

// Command is the base interface for any command
type Command interface {
	Execute(job jobs.Job) (string, error)
	Cmd() string
	HasHandshake() bool
	Templates() map[string]string
	AuthStrategy() string
	AllowedGroups() []string
	Args() []string
	Timeout() time.Duration
	Record() bool
}

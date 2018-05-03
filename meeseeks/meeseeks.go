package meeseeks

import (
	"context"
	"errors"
	"time"
)

// Defaults for commands
const (
	DefaultCommandTimeout = 60 * time.Second
)

// Message interface to interact with an abstract message
type Message interface {
	// The text received without the matching portion
	GetText() string
	// The friendly name of the channel in which the message was issued
	GetChannel() string
	// The channel id used to build the channel link
	GetChannelID() string
	// The channel link is used in replies to show an hyperlink to the channel
	GetChannelLink() string
	// The friendly name of the user that has sent the message, used internally to match with groups and such
	GetUsername() string
	// The username id of the user that has sent the message, used in replies so they notify the user
	GetUserID() string
	// The user link returns a link to the user
	GetUserLink() string
	// IsIM
	IsIM() bool
}

// LoggerProvider wraps the specific logger implementation
type LoggerProvider interface {
	Reader(jobID uint64) LogReader
	Writer(jobID uint64) LogWriter
}

// LogWriter is an interface to write logs to a given job
type LogWriter interface {
	Append(content string) error
	SetError(jobErr error) error
}

// ErrNoLogsForJob is returned when we try to extract the logs of a non existing job
var ErrNoLogsForJob = errors.New("No logs for job")

// LogReader is an interface to read logs from a given job
type LogReader interface {
	// Returns the whole log output of a given job
	Get() (JobLog, error)
	// Head returns the top <limit> log lines
	Head(limit int) (JobLog, error)
	// Tail returns the bottm <limit> log lines
	Tail(limit int) (JobLog, error)
}

// Request is a structure that holds a command execution request
type Request struct {
	Command     string   `json:"Command"`
	Args        []string `json:"Arguments"`
	Username    string   `json:"Username"`
	UserID      string   `json:"UserID"`
	UserLink    string   `json:"UserLink"`
	Channel     string   `json:"Channel"`
	ChannelID   string   `json:"CannelID"`
	ChannelLink string   `json:"CannelLink"`
	IsIM        bool     `json:"IsIM"`
}

// Job represents a request that matched a command and can be executed
type Job struct {
	ID        uint64    `json:"ID"`
	Request   Request   `json:"Request"`
	StartTime time.Time `json:"StartTime"`
	EndTime   time.Time `json:"EndTime"`
	Status    string    `json:"Status"`
}

// JobLog represents all the logging information of a given Job
type JobLog struct {
	Error  string
	Output string
}

// GetError returns nil or an error depending on the current JobLog setup
func (j JobLog) GetError() error {
	if j.Error == "" {
		return nil
	}
	return errors.New(j.Error)
}

// APIToken is a persisted API token pointing to a message used to trigger a command request
type APIToken struct {
	TokenID     string    `json:"token"`
	UserLink    string    `json:"userLink"`
	ChannelLink string    `json:"channelLink"`
	Text        string    `json:"text"`
	CreatedOn   time.Time `json:"created_on"`
}

// Command is the base interface for any command
type Command interface {
	Execute(context.Context, Job) (string, error)
	Cmd() string
	HasHandshake() bool
	Templates() map[string]string
	AuthStrategy() string
	AllowedGroups() []string
	ChannelStrategy() string
	AllowedChannels() []string
	Args() []string
	Timeout() time.Duration
	Help() Help
	Record() bool
}

// Help is the base interface for any command help
type Help interface {
	GetSummary() string
	GetArgs() []string
}

// Alias represent an alias for a command
type Alias struct {
	Alias   string
	Command string
	Args    []string
}

// JobFilter provides the basic tooling to filter jobs when using Find
type JobFilter struct {
	Limit int
	Match func(Job) bool
}

// Jobs status
const (
	JobRunningStatus = "Running"
	JobFailedStatus  = "Failed"
	JobKilledStatus  = "Killed"
	JobSuccessStatus = "Successful"
)

// Jobs provides an interface to handle persistent access to recorded jobs
type Jobs interface {
	// Get returns an existing job by id
	Get(id uint64) (Job, error)

	// Null returns a null job that will not be tracked
	Null(r Request) Job

	// Create records a request in the DB and hands off a new job
	Create(r Request) (Job, error)

	// Fail accounds for the job ending and sets the status.
	Fail(jobID uint64) error

	// Succeed accounds for the job ending and sets the status.
	Succeed(jobID uint64) error

	// Find will walk through the values on the jobs bucket and will apply the Match function
	// to determine if the job matches a search criteria.
	//
	// Returns a list of jobs in descending order that match the filter
	Find(filter JobFilter) ([]Job, error)

	// FailRunningJobs flags as failed any jobs that is still in running state
	FailRunningJobs() error
}

// ErrNoJobWithID is returned when we can't find a job with the proposed id
var ErrNoJobWithID = errors.New("no job could be found")

// APITokens provides an interface to handle persisted api tokens
type APITokens interface {
	// Create creates a new token persistence record and returns the created token.
	Create(userLink, channelLink, text string) (string, error)

	// Get returns the token given an ID, it may return ErrTokenNotFound when there is no such token
	Get(tokenID string) (APIToken, error)

	// Revoke destroys a token by ID
	Revoke(tokenID string) error

	// Find returns a list of tokens that match the filter
	Find(filter APITokenFilter) ([]APIToken, error)
}

// APITokenFilter is used to filter the tokens to be returned from a List query
type APITokenFilter struct {
	Limit int
	Match func(APIToken) bool
}

// Aliases provides an interface to handle persisted aliases
type Aliases interface {
	// Get returns the command for an alias
	Get(userID, alias string) (string, []string, error)

	// List returns all configured aliases for a user ID
	List(userID string) ([]Alias, error)

	// Create adds a new alias for a user ID
	Create(userID, alias, command string, args ...string) error

	// Remove deletes an alias for a user ID
	Remove(userID, alias string) error
}

package persistence

import (
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/persistence/aliases"
	"github.com/gomeeseeks/meeseeks-box/persistence/jobs"
	logs "github.com/gomeeseeks/meeseeks-box/persistence/logs/local"
	"github.com/gomeeseeks/meeseeks-box/persistence/tokens"
)

var provider Provider

func init() {
	provider = Provider{
		Aliases:   aliases.Aliases{},
		Jobs:      jobs.Jobs{},
		APITokens: tokens.Tokens{},
		LogReader: logs.NewReader(),
		LogWriter: logs.NewWriter(),
	}
}

// Provider holds different service implementations to access them, must be initialized
type Provider struct {
	Aliases   meeseeks.Aliases
	Jobs      meeseeks.Jobs
	APITokens meeseeks.APITokens
	LogReader meeseeks.LogReader
	LogWriter meeseeks.LogWriter
}

// Aliases returns an actual instance of the aliases service
func Aliases() meeseeks.Aliases {
	return provider.Aliases
}

// Jobs returns an actual instance of the jobs service
func Jobs() meeseeks.Jobs {
	return provider.Jobs
}

// APITokens returns an actual instance of the APITokens service
func APITokens() meeseeks.APITokens {
	return provider.APITokens
}

// LogReader returns an actual instance of the log reader service
func LogReader() meeseeks.LogReader {
	return provider.LogReader
}

// LogWriter returns an actual instance of the log writer service
func LogWriter() meeseeks.LogWriter {
	return provider.LogWriter
}

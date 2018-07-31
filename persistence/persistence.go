package persistence

import (
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/persistence/aliases"
	"github.com/gomeeseeks/meeseeks-box/persistence/jobs"
	logs "github.com/gomeeseeks/meeseeks-box/persistence/logs/local"
	"github.com/gomeeseeks/meeseeks-box/persistence/tokens"
)

var providers Providers

func init() {
	providers = Providers{
		Aliases:   aliases.Aliases{},
		Jobs:      jobs.Jobs{},
		APITokens: tokens.Tokens{},
		LogReader: logs.NewReader(),
		LogWriter: logs.NewWriter(),
	}
}

// Providers holds different service implementations to access them, must be initialized
type Providers struct {
	Aliases   meeseeks.Aliases
	Jobs      meeseeks.Jobs
	APITokens meeseeks.APITokens
	LogReader meeseeks.LogReader
	LogWriter meeseeks.LogWriter
}

// Aliases returns an actual instance of the aliases service
func Aliases() meeseeks.Aliases {
	return providers.Aliases
}

// Jobs returns an actual instance of the jobs service
func Jobs() meeseeks.Jobs {
	return providers.Jobs
}

// APITokens returns an actual instance of the APITokens service
func APITokens() meeseeks.APITokens {
	return providers.APITokens
}

// LogReader returns an actual instance of the log reader service
func LogReader() meeseeks.LogReader {
	return providers.LogReader
}

// LogWriter returns an actual instance of the log writer service
func LogWriter() meeseeks.LogWriter {
	return providers.LogWriter
}

// Register registers new providers
func Register(proposed Providers) {
	if proposed.Jobs != nil {
		providers.Jobs = proposed.Jobs
	}
	if proposed.LogReader != nil {
		providers.LogReader = proposed.LogReader
	}
	if proposed.LogWriter != nil {
		providers.LogWriter = proposed.LogWriter
	}
}

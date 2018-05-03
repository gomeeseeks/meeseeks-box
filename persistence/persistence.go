package persistence

import (
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/persistence/aliases"
	"github.com/gomeeseeks/meeseeks-box/persistence/jobs"
	logs "github.com/gomeeseeks/meeseeks-box/persistence/logs/provider"
	"github.com/gomeeseeks/meeseeks-box/persistence/tokens"
)

var provider Provider

func init() {
	provider = Provider{
		Aliases:        aliases.Aliases{},
		Jobs:           jobs.Jobs{},
		APITokens:      tokens.Tokens{},
		LoggerProvider: logs.New(logs.LocalLogger),
	}
}

// Provider holds different service implementations to access them, must be initialized
type Provider struct {
	Aliases        meeseeks.Aliases
	Jobs           meeseeks.Jobs
	APITokens      meeseeks.APITokens
	LoggerProvider meeseeks.LoggerProvider
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

// LoggerProvider returns an actual instance of the loggerProvider service
func LoggerProvider() meeseeks.LoggerProvider {
	return provider.LoggerProvider
}

package logs

import (
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
)

var provider meeseeks.LoggerProvider

func must() meeseeks.LoggerProvider {
	if provider == nil {
		panic("Meeseeks logger provider is not initialized")
	}
	return provider
}

// Configure sets the logger provider
func Configure(p meeseeks.LoggerProvider) {
	provider = p
}

// Reader returns a LogReader attached to the jobID
func Reader() meeseeks.LogReader {
	return must().Reader()
}

// Writer returns a LogWriter attached to the jobID
func Writer() meeseeks.LogWriter {
	return must().Writer()
}

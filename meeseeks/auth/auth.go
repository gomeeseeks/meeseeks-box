package auth

import (
	log "github.com/sirupsen/logrus"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
)

// Authorizer is the interface used to check if a user is allowed to run a command
type Authorizer interface {
	IsAllowed(string, config.Command) bool
}

var authStrategies = map[string]Authorizer{
	config.AuthStrategyAny:      anyUserAllowed{},
	config.AuthStrategyUserList: userInListIsAllowed{},
	config.AuthStrategyNone:     noUserAllowed{},
}

// IsAllowed checks if a user is allowed to run a command given the command authorization strategy
func IsAllowed(username string, cmd config.Command) bool {
	strategy, ok := authStrategies[cmd.AuthStrategy]
	if !ok {
		log.Errorf("Command does not have a valid auth strategy, falling back to none: %+v", cmd)
		strategy = authStrategies[config.AuthStrategyNone]
	}
	return strategy.IsAllowed(username, cmd)
}

type anyUserAllowed struct {
}

// IsAllowed implements Authorizer.IsAllowed
func (a anyUserAllowed) IsAllowed(_ string, _ config.Command) bool {
	return true
}

type noUserAllowed struct {
}

// IsAllowed implements Authorizer.IsAllowed
func (a noUserAllowed) IsAllowed(_ string, _ config.Command) bool {
	return false
}

type userInListIsAllowed struct {
}

// IsAllowed implements Authorizer.IsAllowed
func (a userInListIsAllowed) IsAllowed(username string, cmd config.Command) bool {
	for _, u := range cmd.Authorized {
		if username == u {
			return true
		}
	}
	return false
}

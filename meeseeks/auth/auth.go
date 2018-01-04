package auth

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
)

// Authorizer is the interface used to check if a user is allowed to run a command
type Authorizer interface {
	Check(string, config.Command) error
}

// ErrUserNotAllowed is the error returned when the auth check fails
var ErrUserNotAllowed = errors.New("User no allower")

var authStrategies = map[string]Authorizer{
	config.AuthStrategyAny:      anyUserAllowed{},
	config.AuthStrategyUserList: userInListIsAllowed{},
	config.AuthStrategyNone:     noUserAllowed{},
}

// Check checks if a user is allowed to run a command given the command authorization strategy
func Check(username string, cmd config.Command) error {
	strategy, ok := authStrategies[cmd.AuthStrategy]
	if !ok {
		log.Errorf("Command does not have a valid auth strategy, falling back to none: %+v", cmd)
		strategy = authStrategies[config.AuthStrategyNone]
	}
	return strategy.Check(username, cmd)
}

type anyUserAllowed struct {
}

// Check implements Authorizer.Check
func (a anyUserAllowed) Check(_ string, _ config.Command) error {
	return nil
}

type noUserAllowed struct {
}

// Check implements Authorizer.Check
func (a noUserAllowed) Check(_ string, _ config.Command) error {
	return ErrUserNotAllowed
}

type userInListIsAllowed struct {
}

// IsAllowed implements Authorizer.IsAllowed
func (a userInListIsAllowed) Check(username string, cmd config.Command) error {
	for _, u := range cmd.Authorized {
		if username == u {
			return nil
		}
	}
	return ErrUserNotAllowed
}

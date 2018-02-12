package auth

import (
	"errors"

	log "github.com/sirupsen/logrus"
)

// AdminGroup is the name of the admin group
const AdminGroup = "admin"

// CommandAuthorization is the interface for the command authorization
type CommandAuthorization interface {
	AuthStrategy() string
	AllowedGroups() []string
}

// Authorizer is the interface used to check if a user is allowed to run a command
type Authorizer interface {
	Check(string, CommandAuthorization) error
}

// ErrUserNotAllowed is the error returned when the auth check fails
var ErrUserNotAllowed = errors.New("User no allower")

// Authorization Strategies determine who has access to what
const (
	AuthStrategyAny          = "any"
	AuthStrategyAllowedGroup = "group"
	AuthStrategyNone         = "none"
)

var authStrategies = map[string]Authorizer{
	AuthStrategyAny:          anyUserAllowed{},
	AuthStrategyAllowedGroup: userInGroupAllowed{},
	AuthStrategyNone:         noUserAllowed{},
}

// Check checks if a user is allowed to run a command given the command authorization strategy
func Check(username string, cmd CommandAuthorization) error {
	strategy, ok := authStrategies[cmd.AuthStrategy()]
	if !ok {
		log.Errorf("Command does not have a valid auth strategy, falling back to none: %+v", cmd)
		strategy = authStrategies[AuthStrategyNone]
	}
	return strategy.Check(username, cmd)
}

type anyUserAllowed struct {
}

// Check implements Authorizer.Check
func (a anyUserAllowed) Check(_ string, _ CommandAuthorization) error {
	return nil
}

type noUserAllowed struct {
}

// Check implements Authorizer.Check
func (a noUserAllowed) Check(_ string, _ CommandAuthorization) error {
	return ErrUserNotAllowed
}

type userInGroupAllowed struct {
}

func (a userInGroupAllowed) Check(username string, cmd CommandAuthorization) error {
	for _, group := range cmd.AllowedGroups() {
		err := groups.CheckUserInGroup(username, group)
		switch err {
		case nil:
			log.Debugf("User %s found in group %s", username, group)
			return nil
		case ErrUserNotInGroup:
			log.Debugf("User %s is not in group %s", username, group)
		case ErrGroupNotFound:
			log.Errorf("Could not found group %s", group)
		default:
			log.Errorf("Unexpected error %s", err)
		}
	}
	return ErrUserNotAllowed
}

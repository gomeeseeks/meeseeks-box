package auth

import (
	"errors"
	"fmt"
	"sort"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	log "github.com/sirupsen/logrus"
)

// AdminGroup points to the "admin" group, the only hard coded one
const AdminGroup = "admin"

// CommandAuthorization represents the authorization model for a command
type CommandAuthorization interface {
	AuthStrategy() string
	AllowedGroups() []string
	AllowedChannels() []string
}

// Authorizer is the interface used to check if a user is allowed to run a command
type Authorizer interface {
	Check(string, CommandAuthorization) error
}

// ErrUserNotAllowed is the error returned when the auth check fails because the user is not in an allowed group
var ErrUserNotAllowed = errors.New("User no allowed")

// ErrChannelNotAllowed is the error returne when the auth check fails because the command was invoked in a not allowed channel
var ErrChannelNotAllowed = errors.New("Command not allowed in channel")

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
func Check(req meeseeks.Request, cmd CommandAuthorization) error {
	strategy, ok := authStrategies[cmd.AuthStrategy()]
	if !ok {
		log.Errorf("Command does not have a valid auth strategy, falling back to none: %+v", cmd)
		strategy = authStrategies[AuthStrategyNone]
	}
	if err := strategy.Check(req.Username, cmd); err != nil {
		return err
	}

	if len(cmd.AllowedChannels()) == 0 {
		return nil
	}

	for _, ch := range cmd.AllowedChannels() {
		if req.Channel == ch {
			return nil
		}
	}
	return ErrChannelNotAllowed
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

// Groups is used to keep configured groups
type Groups struct {
	groups map[string]map[string]bool
}

var groups *Groups
var knownUsers map[string]struct{}

// Errors
var (
	ErrGroupNotFound  = fmt.Errorf("Groups does not exists")
	ErrUserNotInGroup = fmt.Errorf("User does not belong to group")
)

// Configure loads all the configured groups
//
// This should go away the moment we start storing groups in some storage
func Configure(configuredGroups map[string][]string) {
	g := Groups{
		groups: map[string]map[string]bool{},
	}
	knownUsers = make(map[string]struct{})

	for name, users := range configuredGroups {
		group := make(map[string]bool)
		for _, user := range users {
			group[user] = true
			knownUsers[user] = struct{}{}
		}
		g.groups[name] = group
	}

	groups = &g
}

// CheckUserInGroup returns nil if the user belongs to the given group, else, an error
func (g *Groups) CheckUserInGroup(username, group string) error {
	users, ok := g.groups[group]
	if !ok {
		return ErrGroupNotFound
	}
	if _, ok := users[username]; !ok {
		return ErrUserNotInGroup
	}
	return nil
}

// GetGroups returns the groups and users that are setup
func GetGroups() map[string][]string {
	g := make(map[string][]string)
	for group, users := range groups.groups {
		groupUsers := make([]string, 0)
		for user := range users {
			groupUsers = append(groupUsers, user)
		}
		sort.Strings(groupUsers) // Sort them to be stable
		g[group] = groupUsers
	}
	return g
}

// IsKnownUser returns true if the user is configured in any group
func IsKnownUser(username string) (ok bool) {
	_, ok = knownUsers[username]
	return
}

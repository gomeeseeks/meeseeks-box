package auth

import (
	"errors"
	"fmt"
	"sort"

	"gitlab.com/yakshaving.art/meeseeks-box/meeseeks"
	log "github.com/sirupsen/logrus"
)

// AdminGroup points to the "admin" group, the only hard coded one
const AdminGroup = "admin"

// CommandAuthorization represents the authorization model for a command
type CommandAuthorization interface {
	GetAuthStrategy() string
	GetAllowedGroups() []string
	GetChannelStrategy() string
	GetAllowedChannels() []string
}

// Authorizer is the interface used to check if a user is allowed to run a command
type Authorizer interface {
	Check(meeseeks.Request, CommandAuthorization) error
}

// ErrUserNotAllowed is the error returned when the auth check fails because the user is not in an allowed group
var ErrUserNotAllowed = errors.New("user no allowed")

// ErrChannelNotAllowed is the error returne when the auth check fails because the command was invoked in a not allowed channel
var ErrChannelNotAllowed = errors.New("command not allowed in channel")

// ErrOnlyIMAllowed is the error returned when the auth check fails because the command was invoked on a public channel
var ErrOnlyIMAllowed = errors.New("command only allowed in IM")

// ErrGroupNotFound is the error returned when we authorize a group that is not defined
var ErrGroupNotFound = fmt.Errorf("groups does not exists")

// ErrUserNotInGroup is the error returned when a user does not belong to a given group
var ErrUserNotInGroup = fmt.Errorf("user does not belong to group")

// Authorization Strategies determine who has access to what
const (
	AuthStrategyAny          = "any"
	AuthStrategyAllowedGroup = "group"
	AuthStrategyNone         = "none"
)

// Channel Strategies determine in which kind of channel a command can be executed
const (
	ChannelStrategyAny             = "any"
	ChannelStrategyIMOnly          = "im_only"
	ChannelStrategyAllowedChannels = "channel"
)

var authStrategies = map[string]Authorizer{
	AuthStrategyAny:          anyUserAllowed{},
	AuthStrategyAllowedGroup: userInGroupAllowed{},
	AuthStrategyNone:         noUserAllowed{},
}

var channelStrategies = map[string]Authorizer{
	ChannelStrategyAllowedChannels: channelExplicitlyAllowed{},
	ChannelStrategyAny:             anyChannelAllowed{},
	ChannelStrategyIMOnly:          imOnlyAllowed{},
}

// Check checks if a user is allowed to run a command given the command authorization strategy
func Check(req meeseeks.Request, cmd CommandAuthorization) error {
	authStrategy, ok := authStrategies[cmd.GetAuthStrategy()]
	if !ok {
		log.Errorf("Command does not have a valid auth strategy, falling back to none: %+v", cmd)
		authStrategy = authStrategies[AuthStrategyNone]
	}

	if err := authStrategy.Check(req, cmd); err != nil {
		return err
	}

	channelStrategy, ok := channelStrategies[cmd.GetChannelStrategy()]
	if !ok {
		log.Errorf("Command does not have a valid auth strategy, falling back to any: %+v", cmd)
		channelStrategy = channelStrategies[ChannelStrategyAny]
	}

	return channelStrategy.Check(req, cmd)
}

type anyUserAllowed struct {
}

// Check implements Authorizer.Check
func (a anyUserAllowed) Check(_ meeseeks.Request, _ CommandAuthorization) error {
	return nil
}

type noUserAllowed struct {
}

// Check implements Authorizer.Check
func (a noUserAllowed) Check(_ meeseeks.Request, _ CommandAuthorization) error {
	return ErrUserNotAllowed
}

type userInGroupAllowed struct {
}

func (a userInGroupAllowed) Check(req meeseeks.Request, cmd CommandAuthorization) error {
	username := req.Username

	for _, group := range cmd.GetAllowedGroups() {
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

type anyChannelAllowed struct{}

// Check implements Authorizer.Check
func (a anyChannelAllowed) Check(_ meeseeks.Request, _ CommandAuthorization) error {
	return nil
}

type imOnlyAllowed struct{}

// Check implements Authorizer.Check
func (a imOnlyAllowed) Check(req meeseeks.Request, _ CommandAuthorization) error {
	if req.IsIM {
		return nil
	}
	return ErrOnlyIMAllowed
}

type channelExplicitlyAllowed struct{}

// Check implements Authorizer.Check
func (a channelExplicitlyAllowed) Check(req meeseeks.Request, cmd CommandAuthorization) error {
	for _, ch := range cmd.GetAllowedChannels() {
		if req.Channel == ch {
			return nil
		}
	}
	return ErrChannelNotAllowed
}

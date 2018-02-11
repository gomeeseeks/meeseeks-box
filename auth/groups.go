package auth

import (
	"fmt"
	"sort"
)

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

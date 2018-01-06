package auth

import (
	"fmt"
	"sort"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
)

// Groups is used to keep configured groups
type Groups struct {
	groups map[string]map[string]bool
}

var groups *Groups

// Errors
var (
	ErrGroupNotFound  = fmt.Errorf("Groups does not exists")
	ErrUserNotInGroup = fmt.Errorf("User does not belong to group")
)

// Configure loads all the configured groups
//
// This should go away the moment we start storing groups in some storage
func Configure(cnf config.Config) {
	g := Groups{
		groups: map[string]map[string]bool{},
	}

	for name, users := range cnf.Groups {
		group := make(map[string]bool)
		for _, user := range users {
			group[user] = true
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

package auth

import (
	"fmt"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
)

// Groups is used to keep configured groups
type Groups struct {
	groups map[string]map[string]bool
}

var groups = &Groups{
	groups: map[string]map[string]bool{},
}

// Errors
var (
	ErrGroupNotFound  = fmt.Errorf("Groups does not exists")
	ErrUserNotInGroup = fmt.Errorf("User does not belong to group")
)

// Configure loads all the configured groups
//
// This should go away the moment we start storing groups in some storage
func Configure(cnf config.Config) {
	for name, users := range cnf.Groups {
		group := make(map[string]bool)
		for _, user := range users {
			group[user] = true
		}
		groups.groups[name] = group
	}
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

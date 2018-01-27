package auth_test

import (
	"testing"

	"github.com/pcarranza/meeseeks-box/auth"
	"github.com/pcarranza/meeseeks-box/commands"
	"github.com/pcarranza/meeseeks-box/commands/shell"
	stubs "github.com/pcarranza/meeseeks-box/testingstubs"
)

func Test_Auth(t *testing.T) {
	auth.Configure(map[string][]string{
		auth.AdminGroup: []string{"admin_user"},
	})
	commands.Add("any", shell.New(shell.CommandOpts{
		Cmd:          "any",
		AuthStrategy: auth.AuthStrategyAny,
	}))
	commands.Add("none", shell.New(shell.CommandOpts{
		Cmd:          "none",
		AuthStrategy: auth.AuthStrategyNone,
	}))
	commands.Add("admins", shell.New(shell.CommandOpts{
		Cmd:           "none",
		AuthStrategy:  auth.AuthStrategyAllowedGroup,
		AllowedGroups: []string{auth.AdminGroup},
	}))

	tt := []struct {
		name     string
		username string
		cmd      string
		expected error
	}{
		{
			name:     "any",
			username: "myself",
			cmd:      "any",
			expected: nil,
		},
		{
			name:     "none",
			username: "myself",
			cmd:      "none",
			expected: auth.ErrUserNotAllowed,
		},
		{
			name:     "authorized groups",
			username: "admin_user",
			cmd:      "admins",
			expected: nil,
		},
		{
			name:     "authorized groups with unauthorized user",
			username: "normal_user",
			cmd:      "admins",
			expected: auth.ErrUserNotAllowed,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cmd, ok := commands.Find(tc.cmd)
			stubs.AssertEquals(t, true, ok)
			if actual := auth.Check(tc.username, cmd); actual != tc.expected {
				t.Fatalf("Check failed with %s", actual)
			}
		})
	}
}

func Test_Groups(t *testing.T) {
	auth.Configure(
		map[string][]string{
			auth.AdminGroup: []string{"user1", "user2"},
			"developer":     []string{"user1"},
		},
	)
	stubs.AssertEquals(t,
		map[string][]string{
			"developer":     []string{"user1"},
			auth.AdminGroup: []string{"user1", "user2"},
		},
		auth.GetGroups())
}

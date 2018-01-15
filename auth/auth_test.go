package auth_test

import (
	"fmt"
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/auth"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/commands"
	stubs "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
)

var authConfig = config.Config{
	Commands: map[string]config.Command{
		"any": config.Command{
			Cmd:          "any",
			Type:         config.ShellCommandType,
			AuthStrategy: config.AuthStrategyAny,
		},
		"none": config.Command{
			Cmd:          "none",
			Type:         config.ShellCommandType,
			AuthStrategy: config.AuthStrategyNone,
		},
		"admins": config.Command{
			Cmd:           "none",
			Type:          config.ShellCommandType,
			AuthStrategy:  config.AuthStrategyAllowedGroup,
			AllowedGroups: []string{config.AdminGroup},
		},
	},
	Groups: map[string][]string{
		config.AdminGroup: []string{"admin_user"},
	},
}

func Test_Auth(t *testing.T) {
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

	auth.Configure(authConfig)

	cmds, err := commands.New(authConfig)

	stubs.Must(t, "can't create commands from configuration,", err)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cmd, err := cmds.Find(tc.cmd)
			stubs.Must(t, fmt.Sprintf("can't find command %s", tc.cmd), err)
			if actual := auth.Check(tc.username, cmd); actual != tc.expected {
				t.Fatalf("Check failed with %s", actual)
			}
		})
	}
}

func Test_Groups(t *testing.T) {
	auth.Configure(
		config.Config{
			Groups: map[string][]string{
				config.AdminGroup: []string{"user1", "user2"},
				"developer":       []string{"user1"},
			},
		},
	)
	stubs.AssertEquals(t,
		map[string][]string{
			"developer":       []string{"user1"},
			config.AdminGroup: []string{"user1", "user2"},
		},
		auth.GetGroups())
}

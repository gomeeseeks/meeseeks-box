package auth_test

import (
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/auth"
	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	stubs "gitlab.com/mr-meeseeks/meeseeks-box/testingstubs"
)

func Test_Auth(t *testing.T) {
	auth.Configure(config.Config{
		Groups: map[string][]string{
			"admins": []string{"admin_user"},
		},
	})
	tt := []struct {
		name     string
		username string
		cmd      config.Command
		expected error
	}{
		{
			name:     "any",
			username: "myself",
			cmd: config.Command{
				Cmd:          "echo",
				AuthStrategy: config.AuthStrategyAny,
			},
			expected: nil,
		},
		{
			name:     "none",
			username: "myself",
			cmd: config.Command{
				Cmd:          "echo",
				AuthStrategy: config.AuthStrategyNone,
			},
			expected: auth.ErrUserNotAllowed,
		},
		{
			name:     "authorized groups",
			username: "admin_user",
			cmd: config.Command{
				Cmd:           "echo",
				AllowedGroups: []string{"admins"},
				AuthStrategy:  config.AuthStrategyAllowedGroup,
			},
			expected: nil,
		},
		{
			name:     "authorized groups with unauthorized user",
			username: "normal_user",
			cmd: config.Command{
				Cmd:           "echo",
				AllowedGroups: []string{"admins"},
				AuthStrategy:  config.AuthStrategyAllowedGroup,
			},
			expected: auth.ErrUserNotAllowed,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if actual := auth.Check(tc.username, tc.cmd); actual != tc.expected {
				t.Fatalf("Check failed with %s", actual)
			}
		})
	}
}

func Test_Groups(t *testing.T) {
	auth.Configure(
		config.Config{
			Groups: map[string][]string{
				"admin":     []string{"user1", "user2"},
				"developer": []string{"user1"},
			},
		},
	)
	stubs.AssertEquals(t,
		map[string][]string{
			"admin":     []string{"user1", "user2"},
			"developer": []string{"user1"},
		},
		auth.GetGroups())
}

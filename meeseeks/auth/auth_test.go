package auth_test

import (
	"testing"

	"gitlab.com/mr-meeseeks/meeseeks-box/config"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/auth"
)

func Test_Auth(t *testing.T) {
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
				Authorized:   []string{},
				AuthStrategy: config.AuthStrategyAny,
			},
			expected: nil,
		},
		{
			name:     "none",
			username: "myself",
			cmd: config.Command{
				Cmd:          "echo",
				Authorized:   []string{"myself"},
				AuthStrategy: config.AuthStrategyNone,
			},
			expected: auth.ErrUserNotAllowed,
		},
		{
			name:     "authorized list including user",
			username: "myself",
			cmd: config.Command{
				Cmd:          "echo",
				Authorized:   []string{"myself"},
				AuthStrategy: config.AuthStrategyUserList,
			},
			expected: nil,
		},
		{
			name:     "authorized list excluding user",
			username: "someoneelse",
			cmd: config.Command{
				Cmd:          "echo",
				Authorized:   []string{"myself"},
				AuthStrategy: config.AuthStrategyUserList,
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

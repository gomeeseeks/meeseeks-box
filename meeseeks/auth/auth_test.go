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
		expected bool
	}{
		{
			name:     "any",
			username: "myself",
			cmd: config.Command{
				Cmd:          "echo",
				Authorized:   []string{},
				AuthStrategy: config.AuthStrategyAny,
			},
			expected: true,
		},
		{
			name:     "none",
			username: "myself",
			cmd: config.Command{
				Cmd:          "echo",
				Authorized:   []string{"myself"},
				AuthStrategy: config.AuthStrategyNone,
			},
			expected: false,
		},
		{
			name:     "authorized list including user",
			username: "myself",
			cmd: config.Command{
				Cmd:          "echo",
				Authorized:   []string{"myself"},
				AuthStrategy: config.AuthStrategyUserList,
			},
			expected: true,
		},
		{
			name:     "authorized list without user",
			username: "someoneelse",
			cmd: config.Command{
				Cmd:          "echo",
				Authorized:   []string{"myself"},
				AuthStrategy: config.AuthStrategyUserList,
			},
			expected: false,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if auth.IsAllowed(tc.username, tc.cmd) != tc.expected {
				t.Fatalf("IsAllowed is not as expected %t", tc.expected)
			}
		})
	}
}

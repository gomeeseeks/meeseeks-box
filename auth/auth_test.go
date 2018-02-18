package auth_test

import (
	"testing"

	"github.com/gomeeseeks/meeseeks-box/auth"
	"github.com/gomeeseeks/meeseeks-box/commands"
	"github.com/gomeeseeks/meeseeks-box/commands/shell"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
)

func Test_Auth(t *testing.T) {
	auth.Configure(map[string][]string{
		auth.AdminGroup: {"admin_user"},
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
		req      meeseeks.Request
		expected error
	}{
		{
			name:     "any",
			username: "myself",
			req: meeseeks.Request{
				Command:     "any",
				Channel:     "general",
				ChannelID:   "123",
				ChannelLink: "<#123>",
				Username:    "myself",
				UserID:      "userid",
			},
			expected: nil,
		},
		{
			name:     "none",
			username: "myself",
			req: meeseeks.Request{
				Command:     "none",
				Channel:     "general",
				ChannelID:   "123",
				ChannelLink: "<#123>",
				Username:    "myself",
				UserID:      "userid",
			},
			expected: auth.ErrUserNotAllowed,
		},
		{
			name:     "authorized groups",
			username: "admin_user",
			req: meeseeks.Request{
				Command:     "admins",
				Channel:     "general",
				ChannelID:   "123",
				ChannelLink: "<#123>",
				Username:    "myself",
				UserID:      "userid",
			},
			expected: nil,
		},
		{
			name:     "authorized groups with unauthorized user",
			username: "normal_user",
			req: meeseeks.Request{
				Command:     "admins",
				Channel:     "general",
				ChannelID:   "123",
				ChannelLink: "<#123>",
				Username:    "myself",
				UserID:      "userid",
			},
			expected: auth.ErrUserNotAllowed,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mocks.Must(t, tc.name, mocks.WithTmpDB(func(_ string) {
				cmd, ok := commands.Find(&tc.req)
				mocks.AssertEquals(t, true, ok)
				if actual := auth.Check(tc.username, cmd); actual != tc.expected {
					t.Fatalf("Check failed with %s", actual)
				}
			}))
		})
	}
}

func Test_Groups(t *testing.T) {
	auth.Configure(
		map[string][]string{
			auth.AdminGroup: {"user1", "user2"},
			"developer":     {"user1"},
		},
	)
	mocks.AssertEquals(t,
		map[string][]string{
			"developer":     {"user1"},
			auth.AdminGroup: {"user1", "user2"},
		},
		auth.GetGroups())
}

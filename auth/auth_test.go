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
	commands.Add(
		commands.CommandRegistration{
			Name: "any",
			Cmd: shell.New(shell.CommandOpts{
				Cmd:          "any",
				AuthStrategy: auth.AuthStrategyAny,
			})},
		commands.CommandRegistration{
			Name: "none",
			Cmd: shell.New(shell.CommandOpts{
				Cmd:          "none",
				AuthStrategy: auth.AuthStrategyNone,
			})},
		commands.CommandRegistration{
			Name: "admins",
			Cmd: shell.New(shell.CommandOpts{
				Cmd:           "none",
				AuthStrategy:  auth.AuthStrategyAllowedGroup,
				AllowedGroups: []string{auth.AdminGroup},
			})},
		commands.CommandRegistration{
			Name: "general-channel-only",
			Cmd: shell.New(shell.CommandOpts{
				Cmd:             "none",
				AuthStrategy:    auth.AuthStrategyAny,
				AllowedChannels: []string{"general"},
				ChannelStrategy: "channel",
			})},
		commands.CommandRegistration{
			Name: "im-only", Cmd: shell.New(shell.CommandOpts{
				Cmd:             "none",
				AuthStrategy:    auth.AuthStrategyAny,
				ChannelStrategy: "im_only",
			})})

	tt := []struct {
		name     string
		req      meeseeks.Request
		expected error
	}{
		{
			name: "any",
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
			name: "none",
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
			name: "authorized groups",
			req: meeseeks.Request{
				Command:     "admins",
				Channel:     "general",
				ChannelID:   "123",
				ChannelLink: "<#123>",
				Username:    "admin_user",
				UserID:      "userid",
			},
			expected: nil,
		},
		{
			name: "authorized groups with unauthorized user",
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
		{
			name: "authorized channel ok",
			req: meeseeks.Request{
				Command:     "general-channel-only",
				Channel:     "general",
				ChannelID:   "123",
				ChannelLink: "<#123>",
				Username:    "myself",
				UserID:      "userid",
			},
			expected: nil,
		},
		{
			name: "unauthorized channel errs",
			req: meeseeks.Request{
				Command:     "general-channel-only",
				Channel:     "random",
				ChannelID:   "123",
				ChannelLink: "<#123>",
				Username:    "myself",
				UserID:      "userid",
			},
			expected: auth.ErrChannelNotAllowed,
		},
		{
			name: "im channel only fails on any other",
			req: meeseeks.Request{
				Command:     "im-only",
				Channel:     "random",
				ChannelID:   "123",
				ChannelLink: "<#123>",
				Username:    "myself",
				UserID:      "userid",
				IsIM:        false,
			},
			expected: auth.ErrOnlyIMAllowed,
		},
		{
			name: "im channel only works on IM",
			req: meeseeks.Request{
				Command:     "im-only",
				Channel:     "who-cares",
				ChannelID:   "123",
				ChannelLink: "<#123>",
				Username:    "myself",
				UserID:      "userid",
				IsIM:        true,
			},
			expected: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mocks.Must(t, tc.name, mocks.WithTmpDB(func(_ string) {
				cmd, ok := commands.Find(&tc.req)
				mocks.AssertEquals(t, true, ok)
				if actual := auth.Check(tc.req, cmd); actual != tc.expected {
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

package tokens_test

import (
	"testing"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/persistence/tokens"
)

func Test_TokenLifecycle(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		id, err := tokens.Create(tokens.NewTokenRequest{
			UserLink:    "myuser",
			ChannelLink: "mychannel",
			Text:        "echo hello",
		})
		mocks.Must(t, "could not create token", err)
		if id == "" {
			t.Fatal("create token returned an empty token id(?)")
		}

		tk, err := tokens.Get(id)
		mocks.Must(t, "could not get token back", err)

		mocks.AssertEquals(t, id, tk.TokenID)
		mocks.AssertEquals(t, "myuser", tk.UserLink)
		mocks.AssertEquals(t, "mychannel", tk.ChannelLink)
		mocks.AssertEquals(t, "echo hello", tk.Text)
	})
}

func Test_TokenListing(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		id, err := tokens.Create(tokens.NewTokenRequest{
			Text:        "echo something",
			UserLink:    "myuser",
			ChannelLink: "mychannel",
		})
		mocks.Must(t, "could not create token", err)

		t1, err := tokens.Get(id)
		mocks.Must(t, "could not get token back", err)

		id, err = tokens.Create(tokens.NewTokenRequest{
			Text:        "echo something else",
			UserLink:    "someone_else",
			ChannelLink: "my_other_channel",
		})
		mocks.Must(t, "could not create token", err)

		t2, err := tokens.Get(id)
		mocks.Must(t, "could not get token back", err)

		tt := []struct {
			Name     string
			Filter   tokens.Filter
			Expected []meeseeks.APIToken
		}{
			{
				Name:     "empty list",
				Expected: []meeseeks.APIToken{},
				Filter: tokens.Filter{
					Limit: 0,
				},
			},
			{
				Name:     "filter by username works",
				Expected: []meeseeks.APIToken{t2},
				Filter: tokens.Filter{
					Limit: 5,
					Match: func(tk meeseeks.APIToken) bool {
						return tk.UserLink == t2.UserLink
					},
				},
			},
			{
				Name:     "filter by channel works",
				Expected: []meeseeks.APIToken{t1},
				Filter: tokens.Filter{
					Limit: 5,
					Match: func(tk meeseeks.APIToken) bool {
						return tk.ChannelLink == t1.ChannelLink
					},
				},
			},
		}
		for _, tc := range tt {
			t.Run(tc.Name, func(t *testing.T) {
				token, err := tokens.Find(tc.Filter)
				mocks.Must(t, "failed to list tokens", err)
				mocks.AssertEquals(t, tc.Expected, token)
			})
		}
	})
}

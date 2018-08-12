package tokens_test

import (
	"testing"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/persistence"
	"github.com/gomeeseeks/meeseeks-box/persistence/tokens"
)

func TestGetNonExistingToken(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		_, err := persistence.APITokens().Get("invalid")
		mocks.AssertEquals(t, tokens.ErrTokenNotFound, err)
	})
}

func TestRevokeNonExistingToken(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		err := persistence.APITokens().Revoke("invalid")
		mocks.AssertEquals(t, tokens.ErrTokenNotFound, err)
	})
}
func Test_TokenLifecycle(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		id, err := persistence.APITokens().Create(
			"myuser",
			"mychannel",
			"echo hello",
		)
		mocks.Must(t, "could not create token", err)
		if id == "" {
			t.Fatal("create token returned an empty token id(?)")
		}

		tk, err := persistence.APITokens().Get(id)
		mocks.Must(t, "could not get token back", err)

		mocks.AssertEquals(t, id, tk.TokenID)
		mocks.AssertEquals(t, "myuser", tk.UserLink)
		mocks.AssertEquals(t, "mychannel", tk.ChannelLink)
		mocks.AssertEquals(t, "echo hello", tk.Text)

		mocks.Must(t, "could not revoke token", persistence.APITokens().Revoke(id))
	})
}

func Test_TokenListing(t *testing.T) {
	mocks.WithTmpDB(func(_ string) {
		id, err := persistence.APITokens().Create(
			"echo something",
			"myuser",
			"mychannel",
		)
		mocks.Must(t, "could not create token", err)

		t1, err := persistence.APITokens().Get(id)
		mocks.Must(t, "could not get token back", err)

		id, err = persistence.APITokens().Create(
			"echo something else",
			"someone_else",
			"my_other_channel",
		)
		mocks.Must(t, "could not create token", err)

		t2, err := persistence.APITokens().Get(id)
		mocks.Must(t, "could not get token back", err)

		tt := []struct {
			Name     string
			Filter   meeseeks.APITokenFilter
			Expected []meeseeks.APIToken
		}{
			{
				Name:     "empty list",
				Expected: []meeseeks.APIToken{},
				Filter: meeseeks.APITokenFilter{
					Limit: 0,
				},
			},
			{
				Name:     "filter by username works",
				Expected: []meeseeks.APIToken{t2},
				Filter: meeseeks.APITokenFilter{
					Limit: 5,
					Match: func(tk meeseeks.APIToken) bool {
						return tk.UserLink == t2.UserLink
					},
				},
			},
			{
				Name:     "filter by channel works",
				Expected: []meeseeks.APIToken{t1},
				Filter: meeseeks.APITokenFilter{
					Limit: 5,
					Match: func(tk meeseeks.APIToken) bool {
						return tk.ChannelLink == t1.ChannelLink
					},
				},
			},
		}
		for _, tc := range tt {
			t.Run(tc.Name, func(t *testing.T) {
				token, err := persistence.APITokens().Find(tc.Filter)
				mocks.Must(t, "failed to list tokens", err)
				mocks.AssertEquals(t, tc.Expected, token)
			})
		}
	})
}

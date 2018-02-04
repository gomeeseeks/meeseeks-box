package tokens_test

import (
	"testing"

	stubs "github.com/pcarranza/meeseeks-box/testingstubs"
	"github.com/pcarranza/meeseeks-box/tokens"
)

func Test_TokenLifecycle(t *testing.T) {
	stubs.WithTmpDB(func(_ string) {
		id, err := tokens.Create(tokens.NewTokenRequest{
			Command: "echo",
			User:    "myuser",
			Channel: "mychannel",
			Args:    []string{"hello"},
		})
		stubs.Must(t, "could not create token", err)
		if id == "" {
			t.Fatal("create token returned an empty token id(?)")
		}

		tk, err := tokens.Get(id)
		stubs.Must(t, "could not get token back", err)

		stubs.AssertEquals(t, id, tk.TokenID)
		stubs.AssertEquals(t, "echo", tk.Command)
		stubs.AssertEquals(t, "myuser", tk.User)
		stubs.AssertEquals(t, "mychannel", tk.Channel)
		stubs.AssertEquals(t, []string{"hello"}, tk.Args)
	})
}

func Test_TokenListing(t *testing.T) {
	stubs.WithTmpDB(func(_ string) {
		id, err := tokens.Create(tokens.NewTokenRequest{
			Command: "echo",
			User:    "myuser",
			Channel: "mychannel",
			Args:    []string{"hello"},
		})
		stubs.Must(t, "could not create token", err)

		t1, err := tokens.Get(id)
		stubs.Must(t, "could not get token back", err)

		id, err = tokens.Create(tokens.NewTokenRequest{
			Command: "echo",
			User:    "someone_else",
			Channel: "my_other_channel",
			Args:    []string{"hello"},
		})
		stubs.Must(t, "could not create token", err)

		t2, err := tokens.Get(id)
		stubs.Must(t, "could not get token back", err)

		tt := []struct {
			Name     string
			Filter   tokens.Filter
			Expected []tokens.Token
		}{
			{
				Name:     "empty list",
				Expected: []tokens.Token{},
				Filter: tokens.Filter{
					Limit: 0,
				},
			},
			{
				Name:     "filter by username works",
				Expected: []tokens.Token{t2},
				Filter: tokens.Filter{
					Limit: 5,
					Match: func(tk tokens.Token) bool {
						return tk.User == t2.User
					},
				},
			},
			{
				Name:     "filter by channel works",
				Expected: []tokens.Token{t1},
				Filter: tokens.Filter{
					Limit: 5,
					Match: func(tk tokens.Token) bool {
						return tk.Channel == t1.Channel
					},
				},
			},
		}
		for _, tc := range tt {
			t.Run(tc.Name, func(t *testing.T) {
				token, err := tokens.Find(tc.Filter)
				stubs.Must(t, "failed to list tokens", err)
				stubs.AssertEquals(t, tc.Expected, token)
			})
		}
	})
}

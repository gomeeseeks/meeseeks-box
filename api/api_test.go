package api_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gomeeseeks/meeseeks-box/api"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/mocks"
	"github.com/gomeeseeks/meeseeks-box/persistence"
)

/*
   # API

   ## Will need

   * a token that links to a user+command+channel
   * a command to create a token
   * a command to revoke a token
   * a command to list tokens with commands

   ## Steps to enable the API

   1. Configuration: the endpoint in which to listen, namely address and path.
   2. Enabling: using a command, pass in the command that the API will be
   calling and the channel in which to reply, else it will be considered as an
   IM
   3. The command will need to resolve the channel <token> (the client will need to provide this)
   4. This command can only be called over IM
   5. ~The api enabling will need to test that the user has access to the requested command?~ (nah... late binding)

   ## Token data model

   - Tokens
     - Token (something something UUID hash)
       - UserID
       - ChannelID
       - Command
       - Args

   ### Interface

   - Create{payload} - returns the token
   - Get{Token}
   - List{UserID}

   # Steps to build:

   1. Create the data model
   2. Add the creation and the Get methods
   3. Add the http interface to post the token, return 200 when found
   4. Pipe the message to the stub client, check it works
   5. Finish implementation

*/

func TestAPIServer(t *testing.T) {
	mocks.Must(t, "failed to create a temporary DB", mocks.WithTmpDB(func(dbpath string) {
		// This is necessary because we need to store and then load the token in the DB
		mocks.NewHarness().WithEchoCommand().WithDBPath(dbpath).Load()

		validToken, err := persistence.APITokens().Create(
			"someoneLink",
			"generalLink",
			"echo something",
		)
		mocks.Must(t, "failed to create the token", err)

		s := api.NewServer(mocks.EnricherStub{
			IM: false,
		}, "/api", ":0")
		defer s.Shutdown()

		ch := make(chan meeseeks.Request)
		go s.Listen(ch)

		testSrv := httptest.NewServer(http.HandlerFunc(s.HandlePostToken))

		assertHttpStatus := func(statusCode int) func(t *testing.T, actualStatus string) {
			return func(t *testing.T, actualStatus string) {
				mocks.AssertEquals(t, fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
					actualStatus)
			}
		}
		assertNothing := func(_ *testing.T, _ chan meeseeks.Request) {
		}

		tt := []struct {
			name          string
			reqToken      string
			payload       string
			assertStatus  func(*testing.T, string)
			assertRequest func(*testing.T, chan meeseeks.Request)
		}{
			{
				"invalid token",
				"invalid_token",
				"",
				assertHttpStatus(http.StatusUnauthorized),
				assertNothing,
			},
			{
				"no token",
				"",
				"",
				assertHttpStatus(http.StatusBadRequest),
				assertNothing,
			},
			{
				"valid without payload call",
				validToken,
				"",
				assertHttpStatus(http.StatusAccepted),
				func(t *testing.T, ch chan meeseeks.Request) {
					req := <-ch
					mocks.AssertEquals(t, "echo", req.Command)
					mocks.AssertEquals(t, []string{"something"}, req.Args)
					mocks.AssertEquals(t, "general", req.ChannelID)
					mocks.AssertEquals(t, "name: general", req.Channel)
					mocks.AssertEquals(t, "<#general>", req.ChannelLink)
					mocks.AssertEquals(t, "name: someone", req.Username)
					mocks.AssertEquals(t, "<@someone>", req.UserLink)
					mocks.AssertEquals(t, "someone", req.UserID)
					mocks.AssertEquals(t, false, req.IsIM)
				},
			},
			{
				"valid with payload call",
				validToken,
				"with arguments that will be attached",
				assertHttpStatus(http.StatusAccepted),
				func(t *testing.T, ch chan meeseeks.Request) {
					req := <-ch
					mocks.AssertEquals(t, "echo", req.Command)
					mocks.AssertEquals(t, []string{"something", "with", "arguments", "that", "will", "be", "attached"}, req.Args)
				},
			},
		}
		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				values := make(url.Values)
				if tc.payload != "" {
					values.Add("message", tc.payload)
				}

				req, err := http.NewRequest("POST", testSrv.URL, strings.NewReader(values.Encode()))
				mocks.Must(t, "Could not create request", err)

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Add("TOKEN", tc.reqToken)

				resp, err := testSrv.Client().Do(req)
				mocks.Must(t, "failed to execute request", err)

				tc.assertStatus(t, resp.Status)
				tc.assertRequest(t, ch)
			})
		}
	}))
}

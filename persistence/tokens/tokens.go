package tokens

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/persistence/db"

	"github.com/coreos/bbolt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var tokensBucketKey = []byte("tokens")

// ErrTokenNotFound is returned when a token can't be found given an ID
var ErrTokenNotFound = fmt.Errorf("no token found")

// Tokens implements the Tokens interface with a locally stored tokens
type Tokens struct{}

// Create gets a new token request and creates a token persistence record. It returns the created token.
func (Tokens) Create(userLink, channelLink, text string) (string, error) {
	return create(userLink, channelLink, text)
}

// Get returns the token given an ID, it may return ErrTokenNotFound when there is no such token
func (Tokens) Get(tokenID string) (meeseeks.APIToken, error) {
	return get(tokenID)
}

// Revoke destroys a token by ID
func (Tokens) Revoke(tokenID string) error {
	return revoke(tokenID)
}

// Find returns a list of tokens that match the filter
func (Tokens) Find(filter meeseeks.APITokenFilter) ([]meeseeks.APIToken, error) {
	return find(filter)
}

func create(userLink, channelLink, text string) (string, error) {
	token := uuid.New().String()

	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(tokensBucketKey)
		if err != nil {
			return err
		}

		t := meeseeks.APIToken{
			TokenID:     token,
			UserLink:    userLink,
			ChannelLink: channelLink,
			Text:        text,
			CreatedOn:   time.Now(),
		}
		tb, err := json.Marshal(t)
		if err != nil {
			return fmt.Errorf("could not marshal token: %s", err)
		}

		logrus.Debugf("Creating token %#v", t)
		return bucket.Put([]byte(token), tb)
	})
	return token, err
}

func get(tokenID string) (meeseeks.APIToken, error) {
	var token meeseeks.APIToken
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(tokensBucketKey)
		if bucket == nil {
			return ErrTokenNotFound
		}

		payload := bucket.Get([]byte(tokenID))
		if payload == nil {
			return ErrTokenNotFound
		}

		return json.Unmarshal(payload, &token)
	})
	logrus.Debugf("Returning token %#v with ID %s", token, tokenID)
	return token, err
}

func revoke(tokenID string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(tokensBucketKey)
		if bucket == nil {
			return ErrTokenNotFound
		}
		return bucket.Delete([]byte(tokenID))
	})
}

// Find returns a list of tokens that match the filter
func find(filter meeseeks.APITokenFilter) ([]meeseeks.APIToken, error) {
	if filter.Match == nil {
		filter.Match = func(_ meeseeks.APIToken) bool { return true }
	}

	tokens := make([]meeseeks.APIToken, 0)

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(tokensBucketKey)
		if bucket == nil {
			return nil // an empty list is not an error
		}

		c := bucket.Cursor()
		_, payload := c.First()
		for len(tokens) < filter.Limit && payload != nil {

			t := meeseeks.APIToken{}
			if err := json.Unmarshal(payload, &t); err != nil {
				return err
			}

			if filter.Match(t) {
				tokens = append(tokens, t)
			}
			_, payload = c.Next()
		}
		return nil
	})
	logrus.Debugf("Looking up tokens, found %#v", tokens)
	return tokens, err
}

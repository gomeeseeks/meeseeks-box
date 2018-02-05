package tokens

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"

	"github.com/coreos/bbolt"
	"github.com/pcarranza/meeseeks-box/db"
)

var tokensBucketKey = []byte("tokens")

// ErrTokenNotFound is returned when a token can't be found given an ID
var ErrTokenNotFound = fmt.Errorf("no token found")

// NewTokenRequest is used to create a new token
type NewTokenRequest struct {
	UserID    string
	ChannelID string
	Text      string
}

// Token is a persisted token
type Token struct {
	TokenID   string    `json:"token"`
	UserID    string    `json:"userID"`
	ChannelID string    `json:"channelID"`
	Text      string    `json:"text"`
	CreatedOn time.Time `json:"created_on"`
}

// createUUID has been _honored_ from hashicorp UUID
func createUUID() (string, error) {
	buf := make([]byte, 16) // Maybe make this configurable
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		buf[0:4],
		buf[4:6],
		buf[6:8],
		buf[8:10],
		buf[10:16]), nil
}

// Create gets a new token request and creates a token persistence record. It returns the created token.
func Create(r NewTokenRequest) (string, error) {
	token, err := createUUID()
	if err != nil {
		return "", fmt.Errorf("could not create UUID for token: %s", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(tokensBucketKey)
		if err != nil {
			return err
		}

		t := Token{
			TokenID:   token,
			UserID:    r.UserID,
			ChannelID: r.ChannelID,
			Text:      r.Text,
			CreatedOn: time.Now(),
		}
		tb, err := json.Marshal(t)
		if err != nil {
			return fmt.Errorf("could not marshal token: %s", err)
		}

		return bucket.Put([]byte(token), tb)
	})
	return token, err
}

// Get returns the token given an ID, it may return ErrTokenNotFound when there is no such token
func Get(tokenID string) (Token, error) {
	var token Token
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
	return token, err
}

// Filter is used to filter the tokens to be returned from a List query
type Filter struct {
	Limit int
	Match func(Token) bool
}

// Find returns a list of tokens that match the filter
func Find(filter Filter) ([]Token, error) {
	if filter.Match == nil {
		filter.Match = func(_ Token) bool { return true }
	}

	tokens := make([]Token, 0)

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(tokensBucketKey)
		if bucket == nil {
			return nil // an empty list is not an error
		}

		c := bucket.Cursor()
		_, payload := c.First()
		for len(tokens) < filter.Limit && payload != nil {

			t := Token{}
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
	return tokens, err
}

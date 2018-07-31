package aliases

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"

	bolt "github.com/coreos/bbolt"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/persistence/db"
)

var aliasesBucketKey = []byte("aliases")

// ErrAliasNotFound is returned when an alias can't be found
var ErrAliasNotFound = fmt.Errorf("alias not found")

// Aliases provides the interface to a locally persisted alias
type Aliases struct{}

// Get returns the command for an alias
func (Aliases) Get(userID, alias string) (string, []string, error) {
	return get(userID, alias)
}

// List returns all configured aliases for a user ID
func (Aliases) List(userID string) ([]meeseeks.Alias, error) {
	return list(userID)
}

// Create adds a new alias for a user ID
func (Aliases) Create(userID, alias, command string, args ...string) error {
	return create(userID, alias, command, args...)
}

// Remove deletes an alias for a user ID
func (Aliases) Remove(userID, alias string) error {
	return remove(userID, alias)
}

func create(userID, alias, command string, args ...string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := getAliasesBucket(userID, tx)
		if err != nil {
			return err
		}
		a := meeseeks.Alias{
			Alias:   alias,
			Command: command,
			Args:    args,
		}
		aj, err := json.Marshal(a)
		if err != nil {
			return fmt.Errorf("could not marshal alias: %s", err)
		}
		return bucket.Put([]byte(alias), aj)
	})
}

func remove(userID, alias string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := getAliasesBucket(userID, tx)
		if err != nil {
			return err
		}
		if k := bucket.Get([]byte(alias)); k == nil {
			return fmt.Errorf("alias not found")
		}
		err = bucket.Delete([]byte(alias))
		if err != nil {
			return err
		}

		// Delete the bucket if no more aliases are present
		cur := bucket.Cursor()
		if k, _ := cur.First(); k == nil {
			deleteAliasesBucket(userID, tx)
		}
		return nil
	})
}

func list(userID string) ([]meeseeks.Alias, error) {
	aliases := make([]meeseeks.Alias, 0)
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := getAliasesBucket(userID, tx)
		if err != nil {
			return err
		}
		cur := bucket.Cursor()
		for _, payload := cur.First(); payload != nil; _, payload = cur.Next() {
			a := meeseeks.Alias{}
			json.Unmarshal(payload, &a)
			aliases = append(aliases, a)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return aliases, nil
}

func get(userID, alias string) (string, []string, error) {
	logrus.Debugf("looking up command %s", alias)
	var a meeseeks.Alias
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := getAliasesBucket(userID, tx)
		if err != nil {
			return err
		}
		payload := bucket.Get([]byte(alias))
		if payload == nil {
			return ErrAliasNotFound
		}
		return json.Unmarshal(payload, &a)
	})
	if err != nil {
		return "", nil, err
	}
	return a.Command, a.Args, nil
}

// getAliasesBucket returns the aliases for a user ID
func getAliasesBucket(userID string, tx *bolt.Tx) (*bolt.Bucket, error) {
	aliasesBucket, err := tx.CreateBucketIfNotExists(aliasesBucketKey)
	if err != nil {
		return nil, fmt.Errorf("could not get aliases bucket: %s", err)
	}
	return aliasesBucket.CreateBucketIfNotExists([]byte(userID))
}

// deleteAliasesBucket returns the aliases for a user ID
func deleteAliasesBucket(userID string, tx *bolt.Tx) error {
	aliasesBucket, err := tx.CreateBucketIfNotExists(aliasesBucketKey)
	if err != nil {
		return fmt.Errorf("could not get aliases bucket: %s", err)
	}
	return aliasesBucket.DeleteBucket([]byte(userID))
}

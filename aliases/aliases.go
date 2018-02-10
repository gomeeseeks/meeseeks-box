package aliases

import (
	"fmt"

	bolt "github.com/coreos/bbolt"
	"github.com/gomeeseeks/meeseeks-box/db"
)

var aliasesBucketKey = []byte("aliases")

// Add adds a new alias for a user ID
func Add(userID, alias, command string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := getAliasesBucket(userID, tx)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(alias), []byte(command))
	})
}

// Delete deletes an alias for a user ID
func Delete(userID, alias string) error {
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

// List returns all configured aliases for a user ID
func List(userID string) (map[string]string, error) {
	aliases := make(map[string]string)
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := getAliasesBucket(userID, tx)
		if err != nil {
			return err
		}
		cur := bucket.Cursor()
		for a, c := cur.First(); a != nil; a, c = cur.Next() {
			aliases[string(a[:])] = string(c[:])
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return aliases, nil
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

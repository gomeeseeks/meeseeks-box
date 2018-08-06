package db

import (
	"fmt"
	"os"
	"sync"
	"time"

	"encoding/binary"
	bolt "github.com/coreos/bbolt"
)

var databaseConfig DatabaseConfig
var database *bolt.DB
var mutex = sync.Mutex{}

// DatabaseConfig holds the configuration for the BoltDB database
type DatabaseConfig struct {
	Path    string        `yaml:"path"`
	Timeout time.Duration `yaml:"timeout"`
	Mode    os.FileMode   `yaml:"file_mode"`
}

// Configure loads the required configuration to be able of connecting to a database
func Configure(cnf DatabaseConfig) error {
	mutex.Lock()
	defer mutex.Unlock()

	if database != nil { // Close the database if it is currently open. This is probably candidate for a mutex
		database.Close()
	}

	databaseConfig = cnf
	db, err := open()
	if err != nil {
		return err
	}
	database = db
	return nil
}

// Open opens a new connection to the database
func open() (*bolt.DB, error) {
	return bolt.Open(databaseConfig.Path, databaseConfig.Mode, &bolt.Options{
		Timeout: databaseConfig.Timeout,
	})
}

// IDToBytes renders the DB generated uint64 to BigEndian for clustered write on the B-TREE
func IDToBytes(ID uint64) []byte {
	buffer := make([]byte, 8)
	binary.BigEndian.PutUint64(buffer, ID)
	return buffer
}

// IDFromBytes loads a BigEndian ID back into an uint64
func IDFromBytes(ID []byte) uint64 {
	return binary.BigEndian.Uint64(ID)
}

// WithDB invokes the passed function with a valid DB object
func WithDB(f func(db *bolt.DB) error) error {
	if database == nil {
		return fmt.Errorf("database is not initialized")
	}
	return f(database)
}

// Update invokes the passed function with a valid open read-write transaction
func Update(f func(tx *bolt.Tx) error) error {
	return WithDB(func(db *bolt.DB) error {
		return db.Update(f)
	})
}

// View invokes the passed function with a valid open read transaction
func View(f func(tx *bolt.Tx) error) error {
	return WithDB(func(db *bolt.DB) error {
		return db.View(f)
	})
}

// Create invokes the passed function sending the next sequence Id and the bucket that corresponds to
// passed bucket ID
func Create(bucketID []byte, f func(uint64, *bolt.Bucket) error) error {
	return Update(func(tx *bolt.Tx) error {
		id, bucket, err := NextSequenceFor(bucketID, tx)
		if err != nil {
			return fmt.Errorf("could not get next sequence for %s: %s", string(bucketID), err)
		}
		return f(id, bucket)
	})
}

// NextSequenceFor returns the next available sequence and the bucket for the
// bucket ID within the read-write transaction
func NextSequenceFor(bucketID []byte, tx *bolt.Tx) (uint64, *bolt.Bucket, error) {
	bucket, err := tx.CreateBucketIfNotExists(bucketID)
	if err != nil {
		return 0, nil, err
	}
	sequence, err := bucket.NextSequence()
	if err != nil {
		return 0, nil, err
	}
	return sequence, bucket, nil
}

// Close closes the database
func Close() error {
	if database == nil {
		return nil
	}
	return database.Close()
}

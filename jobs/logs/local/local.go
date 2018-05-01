package local

import (
	"errors"

	"fmt"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/persistence/db"

	bolt "github.com/coreos/bbolt"
)

var logsBucketKey = []byte("logs")
var errorKey = []byte("error")

// ErrNoLogsForJob is returned when we try to extract the logs of a non existing job
var ErrNoLogsForJob = errors.New("No logs for job")

// LogWriter is an interface to write job logs
type LogWriter interface {
	Append(content string) error
	SetError(jobErr error) error
}

// localLogWriter contains all that's needed to create a local log writer
type localLogWriter struct {
	jobID uint64
}

// NewJobLogWriter returns a new LogWriter object that uses Bolt as the backend
func NewJobLogWriter(jobID uint64) LogWriter {
	return localLogWriter{
		jobID: jobID,
	}
}

// Append adds a new line to the logs of the given Job
func (l localLogWriter) Append(content string) error {
	if content == "" {
		return nil
	}
	return db.Update(func(tx *bolt.Tx) error {
		jobBucket, err := getJobBucket(l.jobID, tx)
		if err != nil {
			return fmt.Errorf("could not get job %d bucket: %s", l.jobID, err)
		}

		sequence, err := jobBucket.NextSequence()
		if err != nil {
			return fmt.Errorf("could not get next sequence for job %d: %s", l.jobID, err)
		}

		metrics.LogLinesCount.Inc()

		return jobBucket.Put(db.IDToBytes(sequence), []byte(content))
	})
}

// SetError sets the error message for the given Job
func (l localLogWriter) SetError(jobErr error) error {
	if jobErr == nil {
		return nil
	}
	return db.Update(func(tx *bolt.Tx) error {
		jobBucket, err := getJobBucket(l.jobID, tx)
		if err != nil {
			return fmt.Errorf("could not get job %d bucket: %s", l.jobID, err)
		}
		errorBucket, err := jobBucket.CreateBucketIfNotExists(errorKey)
		if err != nil {
			return fmt.Errorf("could not get error bucket for job %d: %s", l.jobID, err)
		}

		return errorBucket.Put(errorKey, []byte(jobErr.Error()))
	})
}

func getJobBucket(jobID uint64, tx *bolt.Tx) (*bolt.Bucket, error) {
	logsBucket, err := tx.CreateBucketIfNotExists(logsBucketKey)
	if err != nil {
		return nil, fmt.Errorf("could not get logs bucket: %s", err)
	}
	return logsBucket.CreateBucketIfNotExists(db.IDToBytes(jobID))

}

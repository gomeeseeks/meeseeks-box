package logs

import (
	"bytes"
	"fmt"

	bolt "github.com/coreos/bbolt"
	"gitlab.com/mr-meeseeks/meeseeks-box/db"
)

var logsBucketKey = []byte("logs")
var errorKey = []byte("error")

// JobLog represents all the logging information of a given Job
type JobLog struct {
	Error  string
	Output string
}

// Append adds a new line to the logs of the given Job
func Append(jobID uint64, content string) error {
	return db.Update(func(tx *bolt.Tx) error {
		jobBucket, err := getJobBucket(jobID, tx)
		if err != nil {
			return fmt.Errorf("could not get job %d bucket: %s", jobID, err)
		}

		sequence, err := jobBucket.NextSequence()
		if err != nil {
			return fmt.Errorf("could not get next sequence for job %d: %s", jobID, err)
		}

		return jobBucket.Put(db.IDToBytes(sequence), []byte(content))
	})
}

// SetError sets the error message for the given Job
func SetError(jobID uint64, err error) error {
	return db.Update(func(tx *bolt.Tx) error {
		jobBucket, err := getJobBucket(jobID, tx)
		if err != nil {
			return fmt.Errorf("could not get job %d bucket: %s", jobID, err)
		}
		errorBucket, err := jobBucket.CreateBucketIfNotExists(errorKey)
		if err != nil {
			return fmt.Errorf("could not get error bucket for job %d: %s", jobID, err)
		}

		return errorBucket.Put(errorKey, []byte(err.Error()))
	})
}

// Get returns the JobLog for the given Job
func Get(jobID uint64) (JobLog, error) {
	job := &JobLog{}
	err := db.View(func(tx *bolt.Tx) error {
		jobBucket, err := getJobBucket(jobID, tx)
		if err != nil {
			return fmt.Errorf("could not get job %d bucket: %s", jobID, err)
		}
		out := bytes.NewBufferString("")
		c := jobBucket.Cursor()
		_, line := c.First()
		for {
			if line == nil {
				break
			}
			out.Write(line)
			_, line = c.Next()
		}
		job.Output = out.String()

		errorBucket, err := jobBucket.CreateBucketIfNotExists(errorKey)
		if err != nil {
			return fmt.Errorf("could not get error bucket for job %d: %s", jobID, err)
		}
		job.Error = string(errorBucket.Get(errorKey))

		return nil
	})
	return *job, err
}

func getJobBucket(jobID uint64, tx *bolt.Tx) (*bolt.Bucket, error) {
	logsBucket, err := tx.CreateBucketIfNotExists(logsBucketKey)
	if err != nil {
		return nil, fmt.Errorf("could not get logs bucket: %s", err)
	}
	return logsBucket.CreateBucketIfNotExists(db.IDToBytes(jobID))

}

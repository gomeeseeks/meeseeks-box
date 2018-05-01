package logs

import (
	"errors"
	"fmt"
	"strings"

	bolt "github.com/coreos/bbolt"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/persistence/db"
)

var logsBucketKey = []byte("logs")
var errorKey = []byte("error")

// ErrNoLogsForJob is returned when we try to extract the logs of a non existing job
var ErrNoLogsForJob = errors.New("no logs for job")

// Append adds a new line to the logs of the given Job
func Append(jobID uint64, content string) error {
	if content == "" {
		return nil
	}
	return db.Update(func(tx *bolt.Tx) error {
		jobBucket, err := getJobBucket(jobID, tx)
		if err != nil {
			return fmt.Errorf("could not get job %d bucket: %s", jobID, err)
		}

		sequence, err := jobBucket.NextSequence()
		if err != nil {
			return fmt.Errorf("could not get next sequence for job %d: %s", jobID, err)
		}

		metrics.LogLinesCount.Inc()

		return jobBucket.Put(db.IDToBytes(sequence), []byte(content))
	})
}

// SetError sets the error message for the given Job
func SetError(jobID uint64, jobErr error) error {
	if jobErr == nil {
		return nil
	}
	return db.Update(func(tx *bolt.Tx) error {
		jobBucket, err := getJobBucket(jobID, tx)
		if err != nil {
			return fmt.Errorf("could not get job %d bucket: %s", jobID, err)
		}
		errorBucket, err := jobBucket.CreateBucketIfNotExists(errorKey)
		if err != nil {
			return fmt.Errorf("could not get error bucket for job %d: %s", jobID, err)
		}

		return errorBucket.Put(errorKey, []byte(jobErr.Error()))
	})
}

// Get returns the JobLog for the given Job
func Get(jobID uint64) (meeseeks.JobLog, error) {
	job := &meeseeks.JobLog{}
	err := readLogBucket(jobID, func(j *bolt.Bucket) error {
		c := j.Cursor()
		_, line := c.First()
		lines := make([]string, 0)
		for {
			if line == nil {
				break
			}
			lines = append(lines, string(line))
			_, line = c.Next()
		}
		job.Output = strings.Join(lines, "\n")

		errorBucket := j.Bucket(errorKey)
		if errorBucket != nil {
			job.Error = string(errorBucket.Get(errorKey))
		}
		return nil
	})
	return *job, err
}

// Head returns the top <limit> log lines
func Head(jobID uint64, limit int) (meeseeks.JobLog, error) {
	job := &meeseeks.JobLog{}
	err := readLogBucket(jobID, func(j *bolt.Bucket) error {
		c := j.Cursor()

		lines := make([]string, 0)
		_, line := c.First()
		for i := 0; i < limit && line != nil; i++ {
			lines = append(lines, string(line))
			_, line = c.Next()
		}
		job.Output = strings.Join(lines, "\n")

		errorBucket := j.Bucket(errorKey)
		if errorBucket != nil {
			job.Error = string(errorBucket.Get(errorKey))
		}
		return nil
	})
	return *job, err
}

// Tail returns the bottm <limit> log lines
func Tail(jobID uint64, limit int) (meeseeks.JobLog, error) {
	job := &meeseeks.JobLog{}
	err := readLogBucket(jobID, func(j *bolt.Bucket) error {
		c := j.Cursor()
		lines := make([]string, 0)

		_, line := c.Last()
		for i := 0; i < limit && line != nil; i++ {
			lines = append([]string{string(line)}, lines...)
			_, line = c.Prev()
		}
		job.Output = strings.Join(lines, "\n")

		errorBucket := j.Bucket(errorKey)
		if errorBucket != nil {
			job.Error = string(errorBucket.Get(errorKey))
		}
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

func readLogBucket(jobID uint64, f func(*bolt.Bucket) error) error {
	return db.View(func(tx *bolt.Tx) error {
		logsBucket := tx.Bucket(logsBucketKey)
		if logsBucket == nil {
			return ErrNoLogsForJob
		}

		jobBucket := logsBucket.Bucket(db.IDToBytes(jobID))
		if jobBucket == nil {
			return ErrNoLogsForJob
		}

		return f(jobBucket)
	})
}

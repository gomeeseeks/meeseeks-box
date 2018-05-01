package local

import (
	"fmt"
	"strings"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/persistence/db"

	bolt "github.com/coreos/bbolt"
)

// New returns a new local LoggerProvider
func New() meeseeks.LoggerProvider {
	return localLoggerProvider{}
}

type localLoggerProvider struct{}

// Reader Implements LoggerProvider.Reader
func (localLoggerProvider) Reader(jobID uint64) meeseeks.LogReader {
	return localReader{
		jobID: jobID,
	}
}

// Writer Implements LoggerProvider.Writer
func (localLoggerProvider) Writer(jobID uint64) meeseeks.LogWriter {
	return localWriter{
		jobID: jobID,
	}
}

type localWriter struct {
	jobID uint64
}

// Implements LogWriter.Append
func (l localWriter) Append(content string) error {
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

// Implements LogWriter.SetError
func (l localWriter) SetError(jobErr error) error {
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

type localReader struct {
	jobID uint64
}

// Get implements LogReader.Get
func (l localReader) Get() (meeseeks.JobLog, error) {
	job := &meeseeks.JobLog{}
	err := readLogBucket(l.jobID, func(j *bolt.Bucket) error {
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

// Head implements LogReader.Head
func (l localReader) Head(limit int) (meeseeks.JobLog, error) {
	job := &meeseeks.JobLog{}
	err := readLogBucket(l.jobID, func(j *bolt.Bucket) error {
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

// Tail implements LogReader.Tail
func (l localReader) Tail(limit int) (meeseeks.JobLog, error) {
	job := &meeseeks.JobLog{}
	err := readLogBucket(l.jobID, func(j *bolt.Bucket) error {
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

var logsBucketKey = []byte("logs")
var errorKey = []byte("error")

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
			return meeseeks.ErrNoLogsForJob
		}

		jobBucket := logsBucket.Bucket(db.IDToBytes(jobID))
		if jobBucket == nil {
			return meeseeks.ErrNoLogsForJob
		}

		return f(jobBucket)
	})
}
package logs

import (
	"errors"
	"fmt"
	bolt "github.com/coreos/bbolt"

	"github.com/gomeeseeks/meeseeks-box/jobs/logs/local"
	"github.com/gomeeseeks/meeseeks-box/jobs/logs/remote"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/gomeeseeks/meeseeks-box/persistence/db"
	"strings"
)

var loggerConfig LoggerConfig

// JobLogReader is an interface to read job logs
type JobLogReader interface {
	Get() (meeseeks.JobLog, error)
	Head(limit int) (meeseeks.JobLog, error)
	Tail(limit int) (meeseeks.JobLog, error)
}

// LoggerConfig is the configuration for the loggers
type LoggerConfig struct {
	LoggerType string
}

// JobLogWriter is an interface to write job logs
type JobLogWriter interface {
	Append(content string) error
	SetError(jobErr error) error
}

// LocalLogReader contains all that's needed to create a local log reader
type LocalLogReader struct {
	jobID uint64
}

var logsBucketKey = []byte("logs")
var errorKey = []byte("error")

// ErrNoLogsForJob is returned when we try to extract the logs of a non existing job
var ErrNoLogsForJob = errors.New("no logs for job")

// Configure loads the required configuration to set up the loggers
func Configure(cnf LoggerConfig) error {
	loggerConfig = cnf
	return nil
}

// GetJobLogReader returns a new JobLogReader that uses Bolt as the backend
func GetJobLogReader(jobID uint64) JobLogReader {
	return LocalLogReader{
		jobID: jobID,
	}
}

// GetJobLogWriter returns a new JobLogWriter. The backend type depends on the configuration
func GetJobLogWriter(jobID uint64) JobLogWriter {
	if loggerConfig.LoggerType == "remote" {
		return remote.NewJobLogWriter(jobID)
	}
	return local.NewJobLogWriter(jobID)
}

// Get returns the JobLog for the given Job
func (l LocalLogReader) Get() (meeseeks.JobLog, error) {
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

// Head returns the top <limit> log lines
func (l LocalLogReader) Head(limit int) (meeseeks.JobLog, error) {
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

// Tail returns the bottm <limit> log lines
func (l LocalLogReader) Tail(limit int) (meeseeks.JobLog, error) {
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

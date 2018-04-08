package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/gomeeseeks/meeseeks-box/db"
	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/meeseeks/metrics"
	"github.com/sirupsen/logrus"
)

// Jobs status
const (
	RunningStatus = "Running"
	FailedStatus  = "Failed"
	KilledStatus  = "Killed"
	SuccessStatus = "Successful"
)

var jobsBucketKey = []byte("jobs")
var runningJobsBucketKey = []byte("running-jobs")

// ErrNoJobWithID is returned when we can't find a job with the proposed id
var ErrNoJobWithID = errors.New("no job could be found")

// NullJob is used to handle requests that are not recorded
func NullJob(req meeseeks.Request) meeseeks.Job {
	return meeseeks.Job{
		ID:        0,
		Request:   req,
		StartTime: time.Now().UTC(),
		Status:    RunningStatus,
	}
}

// Create registers a new job in running state in the database
func Create(req meeseeks.Request) (meeseeks.Job, error) {
	var job *meeseeks.Job
	err := db.Update(func(tx *bolt.Tx) error {
		jobID, bucket, err := db.NextSequenceFor(jobsBucketKey, tx)
		if err != nil {
			return fmt.Errorf("could not get next sequence for %s: %s", string(jobsBucketKey), err)
		}

		job = &meeseeks.Job{
			ID:        jobID,
			Request:   req,
			StartTime: time.Now().UTC(),
			Status:    RunningStatus,
		}
		logrus.Debugf("Creating job %#v", job)

		runningJobsBucket, err := tx.CreateBucketIfNotExists(runningJobsBucketKey)
		if err != nil {
			return fmt.Errorf("could not create running jobs bucket: %s", err)
		}
		if err = runningJobsBucket.Put(db.IDToBytes(job.ID), []byte(RunningStatus)); err != nil {
			return fmt.Errorf("could not save running job ID %d: %s", jobID, err)
		}

		return save(*job, bucket)
	})
	if err != nil {
		return meeseeks.Job{}, fmt.Errorf("failed to create a job %s", err)
	}
	return *job, nil
}

// Get returns a job by id
func Get(id uint64) (meeseeks.Job, error) {
	job := &meeseeks.Job{}
	err := db.View(func(tx *bolt.Tx) error {
		jobsBucket := tx.Bucket(jobsBucketKey)
		if jobsBucket == nil {
			return ErrNoJobWithID
		}
		payload := jobsBucket.Get(db.IDToBytes(id))
		if payload == nil {
			return ErrNoJobWithID
		}
		return json.Unmarshal(payload, job)
	})
	logrus.Debugf("Returning job %#v for ID %d, err: %s", *job, id, err)
	return *job, err
}

// Finish sets the status of a job to whatever end state if it's current status is running
//
// It also sets the end time of the job
func Finish(jobID uint64, status string) error {
	if !(status == SuccessStatus || status == FailedStatus) {
		return fmt.Errorf("invalid status %s", status)
	}
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobsBucketKey)
		job, err := Get(jobID)
		if err != nil {
			return fmt.Errorf("could not get job with id %d: %s", jobID, err)
		}
		if job.Status != RunningStatus {
			return fmt.Errorf("job is not in running status but %s", job.Status)
		}
		runningJobsBucket := tx.Bucket(runningJobsBucketKey)
		if err = runningJobsBucket.Delete(db.IDToBytes(jobID)); err != nil {
			return fmt.Errorf("could not remove job %d from running list: %s", jobID, err)
		}

		job.EndTime = time.Now().UTC()
		job.Status = status

		difference := job.EndTime.Sub(job.StartTime)
		metrics.TaskDurations.WithLabelValues(job.Request.Command).Observe(difference.Seconds())

		return save(job, bucket)
	})
}

// JobFilter provides the basic tooling to filter jobs when using Find
type JobFilter struct {
	Limit int
	Match func(meeseeks.Job) bool
}

// MultiMatch builds a Match function from a list of Match functions
func MultiMatch(matchers ...func(meeseeks.Job) bool) func(meeseeks.Job) bool {
	return func(job meeseeks.Job) bool {
		for _, matcher := range matchers {
			if !matcher(job) {
				return false
			}
		}
		return true
	}
}

// Find will walk through the values on the jobs bucket and will apply the Match function
// to determine if the job matches a search criteria.
//
// Returns a list of jobs in descending order that match the filter
func Find(filter JobFilter) ([]meeseeks.Job, error) {
	latest := make([]meeseeks.Job, 0)
	matcher := func(job meeseeks.Job) bool {
		return true
	}
	if filter.Match != nil {
		matcher = filter.Match
	}
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobsBucketKey)
		if bucket == nil {
			return nil
		}
		cur := bucket.Cursor()
		_, payload := cur.Last()
		for len(latest) < filter.Limit {
			if payload == nil {
				break
			}

			job := meeseeks.Job{}
			if err := json.Unmarshal(payload, &job); err != nil {
				return fmt.Errorf("failed to load Job payload %s", err)
			}
			if matcher(job) {
				latest = append(latest, job)
			}
			_, payload = cur.Prev()
		}
		return nil
	})
	return latest, err
}

func save(job meeseeks.Job, bucket *bolt.Bucket) error {
	buffer, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return bucket.Put(db.IDToBytes(job.ID), buffer)
}

// FailRunningJobs flags as failed any jobs that is still in running state
func FailRunningJobs() error {
	return db.Update(func(tx *bolt.Tx) error {
		runningJobsBucket := tx.Bucket(runningJobsBucketKey)
		if runningJobsBucket == nil {
			return nil
		}

		jobsBucket := tx.Bucket(jobsBucketKey)
		if jobsBucket == nil {
			return nil
		}

		c := runningJobsBucket.Cursor()
		jobIDKey, _ := c.First()
		for {
			if jobIDKey == nil {
				break
			}
			jobID := db.IDFromBytes(jobIDKey)
			logrus.Warnf("Found job %d in running state, marking as killed", jobID)

			j := meeseeks.Job{}
			if err := json.Unmarshal(jobsBucket.Get(jobIDKey), &j); err != nil {
				return fmt.Errorf("could not read job %d from bucket: %s", jobID, err)
			}

			j.Status = KilledStatus
			j.EndTime = time.Now().UTC()
			if err := save(j, jobsBucket); err != nil {
				return fmt.Errorf("could not save killed job %d: %s", jobID, err)
			}

			if err := runningJobsBucket.Delete(jobIDKey); err != nil {
				return fmt.Errorf("could not delete running job %d: %s", jobID, err)
			}

			jobIDKey, _ = c.Next()
		}
		return nil
	})
}

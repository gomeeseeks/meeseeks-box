package jobs

import (
	"errors"
	"fmt"
	"time"

	"github.com/pcarranza/meeseeks-box/db"
	"github.com/pcarranza/meeseeks-box/meeseeks/request"

	"encoding/json"

	bolt "github.com/coreos/bbolt"
	log "github.com/sirupsen/logrus"
)

// Jobs status
const (
	RunningStatus = "Running"
	FailedStatus  = "Failed"
	SuccessStatus = "Successful"
)

var jobsBucketKey = []byte("jobs")

// ErrNoJobWithID is returned when we can't find a job with the proposed id
var ErrNoJobWithID = errors.New("no job could be found")

// Job represents a single job
type Job struct {
	ID        uint64          `json:"ID"`
	Request   request.Request `json:"Request"`
	StartTime time.Time       `json:"StartTime"`
	EndTime   time.Time       `json:"EndTime"`
	Status    string          `json:"Status"`
}

// NullJob is used to handle requests that are not recorded
func NullJob(req request.Request) Job {
	return Job{
		ID:        0,
		Request:   req,
		StartTime: time.Now().UTC(),
		Status:    RunningStatus,
	}
}

// Create registers a new job in running state in the database
func Create(req request.Request) (Job, error) {
	var job *Job
	err := db.Create(jobsBucketKey, func(jobID uint64, bucket *bolt.Bucket) error {
		job = &Job{
			ID:        jobID,
			Request:   req,
			StartTime: time.Now().UTC(),
			Status:    RunningStatus,
		}

		log.Debugf("Creating job %#v", job)
		return save(job, bucket)
	})
	if err != nil {
		return Job{}, fmt.Errorf("failed to create a job %s", err)
	}
	return *job, nil
}

// Get returns a job by id
func Get(id uint64) (Job, error) {
	job := &Job{}
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
	log.Debugf("Returning job %#v for ID %d, err: %s", *job, id, err)
	return *job, err
}

// Finish sets the status of a job to whatever end state if it's current status is running
//
// It also sets the end time of the job
func (j Job) Finish(status string) error {
	if j.ID == 0 {
		return nil
	}
	if !(status == SuccessStatus || status == FailedStatus) {
		return fmt.Errorf("invalid status %s", status)
	}
	return change(j.ID, func(job *Job) error {
		if job.Status != RunningStatus {
			return fmt.Errorf("job is not in running status")
		}
		job.EndTime = time.Now().UTC()
		job.Status = status
		return nil
	})
}

// JobFilter provides the basic tooling to filter jobs when using Find
type JobFilter struct {
	Limit int
	Match func(Job) bool
}

// MultiMatch builds a Match function from a list of Match functions
func MultiMatch(matchers ...func(Job) bool) func(Job) bool {
	return func(job Job) bool {
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
func Find(filter JobFilter) ([]Job, error) {
	latest := make([]Job, 0)
	matcher := func(job Job) bool {
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

			job := Job{}
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

func change(id uint64, f func(job *Job) error) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobsBucketKey)
		job := &Job{}
		if err := json.Unmarshal(bucket.Get(db.IDToBytes(id)), job); err != nil {
			return fmt.Errorf("could not get job with id %d: %s", id, err)
		}
		if err := f(job); err != nil {
			return fmt.Errorf("could not change job with id %d: %s", id, err)
		}
		return save(job, bucket)
	})
}

func save(job *Job, bucket *bolt.Bucket) error {
	buffer, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return bucket.Put(db.IDToBytes(job.ID), buffer)
}

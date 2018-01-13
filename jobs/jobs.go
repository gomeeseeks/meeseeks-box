package jobs

import (
	"fmt"
	"time"

	"gitlab.com/mr-meeseeks/meeseeks-box/db"
	"gitlab.com/mr-meeseeks/meeseeks-box/meeseeks/request"

	"encoding/json"

	bolt "github.com/coreos/bbolt"
)

// Jobs status
const (
	RunningStatus = "Running"
	FailedStatus  = "Failed"
	SuccessStatus = "Successful"
)

// Job represents a single job
type Job struct {
	ID        uint64          `json:"ID"`
	Request   request.Request `json:"Request"`
	StartTime time.Time       `json:"StartTime"`
	EndTime   time.Time       `json:"EndTime"`
	Status    string          `json:"Status"`
}

var jobsBucketKey = []byte("jobs")
var usersBucketKey = []byte("users")

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
		bucket := tx.Bucket(jobsBucketKey)
		return json.Unmarshal(bucket.Get(db.IDToBytes(id)), job)
	})
	return *job, err
}

// Finish sets the status of a job to whatever end state if it's current status is running
//
// It also sets the end time of the job
func Finish(id uint64, status string) error {
	if !(status == SuccessStatus || status == FailedStatus) {
		return fmt.Errorf("invalid status %s", status)
	}
	return change(id, func(job *Job) error {
		if job.Status != RunningStatus {
			return fmt.Errorf("job is not in running status")
		}
		job.EndTime = time.Now().UTC()
		job.Status = status
		return nil
	})
}

// Latest returns the last N jobs
func Latest(limit int) ([]Job, error) {
	latest := make([]Job, 0)
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobsBucketKey)
		cur := bucket.Cursor()
		_, payload := cur.Last()
		for i := 0; i < limit; i++ {
			if payload == nil {
				break
			}

			job := &Job{}
			if err := json.Unmarshal(payload, job); err != nil {
				return fmt.Errorf("failed to load Job payload %s", err)
			}
			latest = append(latest, *job)
			_, payload = cur.Prev()
		}
		return nil
	})
	return latest, err
}

// Last returns the last job for a given username skipping the commands that are as`skip`
func Last(username string, skip string) (Job, error) {
	var last Job
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobsBucketKey)
		cur := bucket.Cursor()
		_, payload := cur.Last()
		for {
			if payload == nil {
				return fmt.Errorf("could find last job for user %s", username)
			}
			job := &Job{}
			if err := json.Unmarshal(payload, job); err != nil {
				return fmt.Errorf("failed to load Job payload %s", err)
			}
			if job.Request.Username == username && job.Request.Command != skip {
				last = *job
				return nil
			}
			_, payload = cur.Prev()
		}
	})
	return last, err
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

package agent

import (
	"context"
	"time"

	"github.com/gomeeseeks/meeseeks-box/meeseeks"
	"github.com/gomeeseeks/meeseeks-box/remote/api"
)

type remoteJobs struct {
	client         api.CommandPipelineClient
	timeoutSeconds time.Duration
}

func (r remoteJobs) Create(meeseeks.Request) (meeseeks.Job, error) {
	return meeseeks.Job{}, nil
}

func (r remoteJobs) Succeed(jobID uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeoutSeconds)
	defer cancel()

	_, err := r.client.Succeed(ctx, &api.Command{JobID: jobID})
	return err
}

func (r remoteJobs) Fail(jobID uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeoutSeconds)
	defer cancel()

	_, err := r.client.Fail(ctx, &api.Command{JobID: jobID})
	return err
}

func (r remoteJobs) FailRunningJobs() error {
	return nil
}

func (r remoteJobs) Find(meeseeks.JobFilter) ([]meeseeks.Job, error) {
	return []meeseeks.Job{}, nil
}

func (r remoteJobs) Get(uint64) (meeseeks.Job, error) {
	return meeseeks.Job{}, nil
}

func (r remoteJobs) Null(meeseeks.Request) meeseeks.Job {
	return meeseeks.Job{}
}

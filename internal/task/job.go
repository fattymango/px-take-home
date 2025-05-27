package task

import (
	"context"
	"fmt"
	"sync"

	"github.com/fattymango/px-take-home/model"
)

type Job struct {
	task   *model.Task
	ctx    context.Context
	cancel context.CancelFunc
}

func NewJob(task *model.Task) *Job {
	ctx, cancel := context.WithCancel(context.TODO())
	return &Job{
		task:   task,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (j *Job) Cancel() {
	j.cancel()
}

type JobCache interface {
	GetJob(id uint64) (*Job, error)
	SetJob(id uint64, job *Job)
	DeleteJob(id uint64)
	GetAllJobs() ([]*Job, error)
}

type InMemoryJobCache struct {
	cache *sync.Map
}

func NewInMemoryJobCache() *InMemoryJobCache {
	return &InMemoryJobCache{
		cache: &sync.Map{},
	}
}

func (c *InMemoryJobCache) GetJob(taskID uint64) (*Job, error) {
	job, ok := c.cache.Load(taskID)
	if !ok {
		return nil, fmt.Errorf("job not found")
	}

	return job.(*Job), nil
}

func (c *InMemoryJobCache) SetJob(taskID uint64, job *Job) {
	c.cache.Store(taskID, job)
}

func (c *InMemoryJobCache) DeleteJob(taskID uint64) {
	c.cache.Delete(taskID)
}

func (c *InMemoryJobCache) GetAllJobs() ([]*Job, error) {
	var jobs []*Job
	c.cache.Range(func(key, value interface{}) bool {
		jobs = append(jobs, value.(*Job))
		return true
	})

	return jobs, nil
}

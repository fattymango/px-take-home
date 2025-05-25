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
	wg     *sync.WaitGroup
	done   chan struct{}
}

func NewJob(wg *sync.WaitGroup, task *model.Task) *Job {
	ctx, cancel := context.WithCancel(context.TODO())
	return &Job{
		task:   task,
		ctx:    ctx,
		cancel: cancel,
		wg:     wg,
		done:   make(chan struct{}),
	}
}
func (j *Job) Cancel() {
	j.cancel()
}
func (j *Job) Done() {
	j.wg.Done()
	close(j.done)
}

func (j *Job) Wait() {
	<-j.done
	fmt.Printf("job #%d finished\n", j.task.ID)
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

func (c *InMemoryJobCache) GetJob(id uint64) (*Job, error) {
	job, ok := c.cache.Load(id)
	if !ok {
		return nil, fmt.Errorf("job not found")
	}

	return job.(*Job), nil
}

func (c *InMemoryJobCache) SetJob(id uint64, job *Job) {
	fmt.Printf("setting job #%d\n", id)
	c.cache.Store(id, job)
}

func (c *InMemoryJobCache) DeleteJob(id uint64) {
	c.cache.Delete(id)
}

func (c *InMemoryJobCache) GetAllJobs() ([]*Job, error) {
	var jobs []*Job
	c.cache.Range(func(key, value interface{}) bool {
		jobs = append(jobs, value.(*Job))
		return true
	})

	return jobs, nil
}

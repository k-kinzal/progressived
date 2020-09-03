package controller

import (
	"context"
	"reflect"
	"runtime"
	"sync"
	"time"
)

type JobFunc func()

type Job struct {
	scheduleTime time.Time
	jobFunc      JobFunc
}

type Scheduler struct {
	mu   sync.Mutex
	jobs []*Job
}

func (s *Scheduler) Add(scheduleTime time.Time, jobFunc JobFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	index := -1
	newJob := &Job{
		scheduleTime: scheduleTime,
		jobFunc:      jobFunc,
	}
	for i, job := range s.jobs {
		if scheduleTime.Before(job.scheduleTime) || scheduleTime.Equal(job.scheduleTime) {
			index = i
			break
		}
	}
	switch index {
	case -1:
		s.jobs = append(s.jobs, newJob)
	default:
		s.jobs = append(s.jobs[:index+1], s.jobs[index:]...)
		s.jobs[index] = newJob
	}
}

func (s *Scheduler) Remove(scheduleTime time.Time, jobFunc JobFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, job := range s.jobs {
		if scheduleTime.Before(job.scheduleTime) || scheduleTime.Equal(job.scheduleTime) {
			if runtime.FuncForPC(reflect.ValueOf(jobFunc).Pointer()).Name() == runtime.FuncForPC(reflect.ValueOf(job.jobFunc).Pointer()).Name() {
				s.jobs = append(s.jobs[:i], s.jobs[i:]...)
			}
			break
		}
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			next := s.jobs[0]
			if t.Before(next.scheduleTime) || t.Equal(next.scheduleTime) {
				next.jobFunc()
				s.Remove(next.scheduleTime, next.jobFunc)
			}
		}
	}
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		mu:   sync.Mutex{},
		jobs: make([]*Job, 0),
	}
}

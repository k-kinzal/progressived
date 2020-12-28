package looper

import (
	"context"
	"golang.org/x/sync/errgroup"
	"sync"
	"time"
)

type Looper struct {
	wg sync.WaitGroup
	doneCh chan struct{}
	sem chan struct{}
}

func (l *Looper) Run(interval time.Duration, fn func() error) {
	l.wg.Add(1)
	defer l.wg.Done()

	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-l.doneCh:
			return
		case <-t.C:
			l.sem <- struct{}{}
			_ = fn()
			<-l.sem
		}
	}
}

func (l *Looper) Shutdown(ctx context.Context) error {
	t := time.NewTicker(100 * time.Millisecond)
	defer t.Stop()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		l.doneCh <- struct{}{}
		l.wg.Wait()

		return nil
	})

	return eg.Wait()
}

func New(concurrency int) *Looper {
	return &Looper{
		wg:     sync.WaitGroup{},
		doneCh: make(chan struct{}, 1),
		sem:    make(chan struct{}, concurrency),
	}
}

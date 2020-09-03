package controller

import (
	"context"
	"github.com/cenkalti/backoff/v4"
	"github.com/k-kinzal/progressived/pkg/logger"
	"github.com/k-kinzal/progressived/pkg/progressived"
	"time"
)

type Controller struct {
	progressived *progressived.Progressived
	scheduler    *Scheduler
	backoff      backoff.BackOff
	interval     time.Duration
	logger       logger.Logger
}

func (c *Controller) rollback() {
	name := c.progressived.TargetName()
	pcr, err := c.progressived.CurrentPercentage()
	if err != nil {
		c.logger.WithField("action", "rollback").Error(err)
		c.scheduler.Add(time.Now().Add(c.backoff.NextBackOff()), c.rollback)
		return
	}
	scheduledPcr, err := c.progressived.PreviousPercentage()
	if err != nil {
		c.logger.WithField("action", "rollback").Error(err)
		c.scheduler.Add(time.Now().Add(c.backoff.NextBackOff()), c.rollback)
		return
	}

	c.logger.WithField("action", "rollback").Infof("rollback from `%f` to `%f` for `%s`", pcr, scheduledPcr, name)
	newPcr, err := c.progressived.Rollback()
	if err != nil {
		switch err.(type) {
		case progressived.AlreadyCompletedError:
			c.logger.WithField("action", "rollback").Warnf("rollback for `%s` is already complete", name)
			c.scheduler.Add(time.Now().Add(c.backoff.NextBackOff()), c.rollback)
		default:
			c.logger.WithField("action", "rollback").Error(err)
			c.scheduler.Add(time.Now().Add(c.backoff.NextBackOff()), c.rollback)
		}
		return
	}
	newScheduledPcr, err := c.progressived.PreviousPercentage()
	scheduleTime := time.Now().Add(c.interval)

	c.logger.WithField("action", "rollback").Infof("next scheduled rollback will be `%f` to `%f` for `%s` at `%s`", newPcr, newScheduledPcr, name, scheduleTime.Format(time.RFC3339))
	c.scheduler.Add(scheduleTime, c.rollback)
}

func (c *Controller) update() {
	name := c.progressived.TargetName()
	pcr, err := c.progressived.CurrentPercentage()
	if err != nil {
		c.logger.WithField("action", "update").Error(err)
		c.scheduler.Add(time.Now().Add(c.backoff.NextBackOff()), c.rollback)
		return
	}
	scheduledPcr, err := c.progressived.PreviousPercentage()
	if err != nil {
		c.logger.WithField("action", "update").Error(err)
		c.scheduler.Add(time.Now().Add(c.backoff.NextBackOff()), c.rollback)
		return
	}

	c.logger.WithField("action", "update").Infof("update from `%f` to `%f` for `%s`", pcr, scheduledPcr, name)
	newPcr, err := c.progressived.Update()
	if err != nil {
		switch err.(type) {
		case progressived.NotMatchMetricsError:
			c.logger.WithField("action", "update").Error(err)
			c.backoff.Reset()
			c.scheduler.Add(time.Now(), c.rollback)
		case progressived.AlreadyCompletedError:
			c.logger.WithField("action", "update").Warnf("update for `%s` is already complete", name)
			c.scheduler.Add(time.Now().Add(c.backoff.NextBackOff()), c.update)
		default:
			c.logger.WithField("action", "update").Error(err)
			c.scheduler.Add(time.Now().Add(c.backoff.NextBackOff()), c.update)
		}
		return
	}
	newScheduledPcr, err := c.progressived.PreviousPercentage()
	scheduleTime := time.Now().Add(c.interval)

	c.logger.WithField("action", "rollback").Infof("next scheduled update will be `%f` to `%f` for `%s` at `%s`", newPcr, newScheduledPcr, name, scheduleTime.Format(time.RFC3339))
	c.scheduler.Add(scheduleTime, c.update)
}

func (c *Controller) Run(ctx context.Context) {
	c.scheduler.Add(time.Now().Add(c.interval), c.update)
	c.scheduler.Start(ctx)
}

func NewController(prog *progressived.Progressived, interval time.Duration, logger logger.Logger) *Controller {
	return &Controller{
		progressived: prog,
		scheduler:    NewScheduler(),
		backoff:      backoff.NewExponentialBackOff(),
		interval:     interval,
		logger:       logger,
	}
}

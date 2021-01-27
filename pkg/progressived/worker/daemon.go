package worker

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/k-kinzal/progressived/pkg/progressived/persistence"
	"time"
)

type DeploymentStream interface {
	Changes() <-chan persistence.Deployment
}

type Daemon struct {
	interval time.Duration
	reader   persistence.Deployments
	sem      chan struct{}

	awsSession *session.Session
}

func (d *Daemon) Start(ctx context.Context) error {
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case change := <-d.reader.Changes():
			d.sem <- struct{}{}
			d.reconcile(change)
			<-d.sem
		case <-ticker.C:
			entities, err := d.reader.Seq()
			if err != nil {
				continue
			}
			for _, entity := range entities {
				go func(entity persistence.Deployment) {
					d.sem <- struct{}{}
					d.reconcile(entity)
					<-d.sem
				}(entity)
			}

		}
	}
}

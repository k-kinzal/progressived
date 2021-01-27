package worker

import (
	"github.com/k-kinzal/progressived/pkg/progressived/persistence"
	"github.com/k-kinzal/progressived/pkg/provider"
)

func newProviderWithEntity(entity persistence.Deployment) (provider.Provider, error) {

}

func (d *Daemon) updateWeight(entity persistence.Deployment) error {
	p, err := newProviderWithEntity(entity)
	if err != nil {
		return err
	}
	weight := entity.Weight()
	currentWeight, err := p.Get()
	if err != nil {
		return err
	}
	if weight == currentWeight {
		return nil
	}
	if err := p.Update(weight); err != nil {
		return err
	}
	return nil
}

func (d *Daemon) reconcile(entity persistence.Deployment) {
	switch entity.Status() {
	case persistence.ReadyStateStatus:
		// to progress
		newEntity, err := entity.Start(false)
		if err != nil {
			switch err.(type) {
			case *persistence.NotScheduledTimeYet:
			default:
				d.logger.Error("")
			}
			return
		}
		if err := d.reader.Put(newEntity); err != nil {
			d.logger.Error(err)
		}
	case "progress":
		// to progress or completed
		newEntity, err := entity.NextStep()
		if err == nil {
			if err := d.reader.Put(newEntity); err != nil {
				d.logger.Error(err)
			}
		}
		switch err.(type) {
		case *persistence.NotScheduledTimeYet:
			if err := d.updateWeight(entity); err != nil {
				d.logger.Error(err)
			}
		default:
			d.logger.Error("")
		}
		return
	case "completed":
		if err := d.updateWeight(entity); err != nil {
			d.logger.Error(err)
		}
	case "rollback":
		newEntity, err := entity.RollbackStep()
		if err == nil {
			if err := d.reader.Put(newEntity); err != nil {
				d.logger.Error(err)
			}
		}
		switch err.(type) {
		case *persistence.NotScheduledTimeYet:
			if err := d.updateWeight(entity); err != nil {
				d.logger.Error(err)
			}
		default:
			d.logger.Error("")
		}
		// to rollback or rollback_completed
	case "rollback_completed":
		// none
		if err := d.updateWeight(entity); err != nil {
			d.logger.Error(err)
		}
	case "pause":
		// to progress or rollback
	default:
		panic("")
	}
}

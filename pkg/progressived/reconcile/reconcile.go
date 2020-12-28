package reconcile

import (
	"github.com/k-kinzal/progressived/pkg/algorithm"
	"github.com/k-kinzal/progressived/pkg/logger"
	"github.com/k-kinzal/progressived/pkg/provider"
	"github.com/k-kinzal/progressived/pkg/storage"
	"time"
)

type Reconciler struct {
	logger logger.Logger
	storage storage.Storage
	algorithm algorithm.Algorithm
	provider provider.Provider
}

func (r *Reconciler) Reconcile() error {
	now := time.Now()
	state, err := r.storage.Read()
	if err != nil {
		r.logger.Errorf("reconcile: %v", err)
		return err
	}

	switch state.Status.Status {
	case storage.StatusReady:
		weight := r.algorithm.Next(state.Spec.Weight)
		r.logger.Debug("reconcile: \"%s\" -> \"%s\": change the weight from \"%f\" to \"%f\" in state", storage.StatusReady, storage.StatusInProgress, state.Spec.Weight, weight)
		state.Spec.Weight = weight
		state.Status.Status = storage.StatusInProgress
		state.Status.UpdateAt = now
		return r.storage.Write(state)
	case storage.StatusPause:
		r.logger.Debug("reconcile: \"%s\": Did not reconcile because \"%s\" is paused", storage.StatusPause, state.Name)
		return nil
	case storage.StatusInProgress:
		if now.Before(state.Status.UpdateAt.Add(state.Spec.Interval)) {
			weight := r.algorithm.Next(state.Spec.Weight)
			if state.Spec.Weight == 0.0 || state.Spec.Weight == 100.0 {
				r.logger.Debug("reconcile: \"%s\" -> \"%s\": change the weight from \"%f\" to \"%f\" in state", storage.StatusInProgress, storage.StatusCompleted, state.Spec.Weight, weight)
				state.Status.Status = storage.StatusCompleted
			} else {
				r.logger.Debug("reconcile: \"%s\": change the weight from \"%f\" to \"%f\" in state", storage.StatusInProgress, state.Spec.Weight, weight)
				state.Status.Status = storage.StatusInProgress
			}
			state.Spec.Weight = weight
			state.Status.UpdateAt = now
			return r.storage.Write(state)
		}
	case storage.StatusRollback:
		if now.Before(state.Status.UpdateAt.Add(state.Spec.Interval)) {
			weight := r.algorithm.Previous(state.Spec.Weight)
			if state.Spec.Weight == 0.0 || state.Spec.Weight == 100.0 {
				r.logger.Debug("reconcile: \"%s\" -> \"%s\": change the weight from \"%f\" to \"%f\" in state", storage.StatusRollback, storage.StatusRollbackCompleted, state.Spec.Weight, weight)
				state.Status.Status = storage.StatusRollbackCompleted
			} else {
				r.logger.Debug("reconcile: \"%s\": change the weight from \"%f\" to \"%f\" in state", storage.StatusRollback, state.Spec.Weight, weight)
				state.Status.Status = storage.StatusRollback
			}
			state.Spec.Weight = weight
			state.Status.UpdateAt = now
			return r.storage.Write(state)
		}
	case storage.StatusRollbackCompleted:
		r.logger.Debug("reconcile: \"%s\": Did not reconcile because \"%s\" is completed ", storage.StatusRollbackCompleted, state.Name)
		return nil
	case storage.StatusCompleted:
		r.logger.Debugf("reconcile: \"%s\": Did not reconcile because \"%s\" is completed", storage.StatusCompleted, state.Name)
		return nil
	}

	weight, err := r.provider.Get()
	if err != nil {
		return err
	}
	if weight == state.Spec.Weight {
		return nil
	}

	r.logger.Infof("reconcile: \"%s\": change the weight from \"%f\" to \"%f\"", state.Status.Status, weight, state.Spec.Weight)
	return r.provider.Update(weight)
}

func Func() error {
	return nil
}

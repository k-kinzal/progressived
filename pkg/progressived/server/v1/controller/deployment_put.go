package controller

import (
	"fmt"
	"github.com/k-kinzal/progressived/pkg/progressived/persistence"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/request"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/response"
)

// PUT /api/v1/deployment/:name
func (c *Controller) PutDeployment(name string, req *request.PutDeploymentRequest) (*response.PutDeploymentResponse, error) {
	provider, err := c.newProviderWithPutDeploymentRequest(req)
	if err != nil {
		return nil, &response.BadRequestError{Messsage: err.Error()}
	}
	stepBehavior, err := c.newStepBehaviorWithPutDeploymentRequest(req)
	if err != nil {
		return nil, &response.BadRequestError{Messsage: err.Error()}
	}
	rollbackBehavior, err := c.newRollbackBehaviorWithPutDeploymentRequest(req)
	if err != nil {
		return nil, &response.BadRequestError{Messsage: err.Error()}
	}
	metrics, err := c.newMetricsWithPutDeploymentRequest(req)
	if err != nil {
		return nil, &response.BadRequestError{Messsage: err.Error()}
	}

	entity, err := c.per.Get(name)
	if err != nil {
		switch err.(type) {
		case *persistence.DeploymentNotFountdError:
			e, err := c.fact.New(name, req.Interval, provider, stepBehavior, rollbackBehavior, metrics)
			if err != nil {
				return nil, &response.BadRequestError{Messsage: err.Error()}
			}
			entity = e
		default:
			return nil, &response.InternalServerError{Err: err}
		}
	} else {
		e, err := entity.Update(req.Interval, provider, stepBehavior, rollbackBehavior, metrics)
		if err != nil {
			return nil, &response.BadRequestError{Messsage: err.Error()}
		}
		entity = e
	}

	if err := c.per.Put(entity); err != nil {
		return nil, &response.InternalServerError{Err: err}
	}

	return response.NewPutDeploymentResponseWith(entity), nil
}
func (c *Controller) newProviderWithPutDeploymentRequest(req *request.PutDeploymentRequest) (persistence.Provider, error) {
	var provider persistence.Provider
	switch persistence.ProviderType(req.Provider.ProviderType) {
	case persistence.Route53ProviderType:
		var hostedZoneId = req.Provider.HostedZoneID
		var recordName = req.Provider.RecordName
		var recordType = req.Provider.RecordType
		var setIdentifier = req.Provider.SetIdentifier
		p, err := c.fact.NewRoute53ProviderSpec(hostedZoneId, recordName, recordType, setIdentifier)
		if err != nil {
			return nil, err
		}
		provider = p
	case persistence.InMemoryProviderType:
		p, err := c.fact.NewInMemoryProviderSpec()
		if err != nil {
			return nil, err
		}
		provider = p
	default:
		return nil, fmt.Errorf("")
	}
	return provider, nil
}

func (c *Controller) newStepBehaviorWithPutDeploymentRequest(req *request.PutDeploymentRequest) (persistence.StepBehavior, error) {
	var stepBehavior persistence.StepBehavior
	switch persistence.BehaviorAlgorithm(req.Step.Algorithm) {
	case persistence.IncreaseBehaviorType:
		s, err := c.fact.NewIncreaseStepBehaviorSpec(req.Step.Threshold)
		if err != nil {
			return nil, err
		}
		stepBehavior = s
	default:
		return nil, fmt.Errorf("")
	}

	return stepBehavior, nil
}

func (c *Controller) newRollbackBehaviorWithPutDeploymentRequest(req *request.PutDeploymentRequest) (persistence.RollbackBehavior, error) {
	var rollbackBehavior persistence.RollbackBehavior
	switch persistence.RollbackBehaviorAlgorithm(req.Rollback.Algorithm) {
	case persistence.RollbackDecreaseBehaviorType:
		r, err := c.fact.NewDecreaseRollbackBehaviorSpec(req.Step.Threshold)
		if err != nil {
			return nil, err
		}
		rollbackBehavior = r
	case persistence.RollbackHistoryBehaviorType:
		r, err := c.fact.NewHistoryRollbackBehaviorSpec()
		if err != nil {
			return nil, err
		}
		rollbackBehavior = r
	default:
		return nil, nil
	}

	return rollbackBehavior, nil
}

func (c *Controller) newMetricsWithPutDeploymentRequest(req *request.PutDeploymentRequest) ([]persistence.Metrics, error) {
	metrics := make([]persistence.Metrics, len(req.Metrics))
	for i, v := range req.Metrics {
		switch persistence.MetricsType(v.MetricType) {
		case persistence.InMemoryMetricsType:
			m, err := c.fact.NewInMemoryMetrics(v.Period, v.Condition, v.Target.Percentage, v.Target.TimeWindow)
			if err != nil {
				return nil, err
			}
			metrics[i] = m
		case persistence.InCloudWatchMetricsType:
			m, err := c.fact.NewCloudWatchMetrics(v.Period, v.Condition, v.Query, *v.AllowNoData, v.Target.Percentage, v.Target.TimeWindow)
			if err != nil {
				return nil, err
			}
			metrics[i] = m
		default:
			return nil, fmt.Errorf("")
		}
	}

	return metrics, nil
}

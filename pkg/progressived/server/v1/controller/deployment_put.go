package controller

import (
	"fmt"
	"github.com/k-kinzal/progressived/pkg/progressived/persistence/v1/deployment"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/request"
	"github.com/k-kinzal/progressived/pkg/progressived/server/v1/response"
	"time"
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

	entity, err := c.deployments.Get(name)
	if err != nil {
		switch err.(type) {
		case *deployment.NotFoundError:
			e, err := deployment.New(name, req.Interval, provider, stepBehavior, rollbackBehavior, metrics)
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
	if req.AutoStart {
		entity, err = entity.Scheduling(nil)
		if err != nil {
			return nil, &response.BadRequestError{Messsage: err.Error()}
		}
	}

	if err := c.deployments.Put(entity); err != nil {
		return nil, &response.InternalServerError{Err: err}
	}

	return response.NewPutDeploymentResponse(entity), nil
}
func (c *Controller) newProviderWithPutDeploymentRequest(req *request.PutDeploymentRequest) (*deployment.ProviderSpec, error) {
	var provider *deployment.ProviderSpec
	switch deployment.ProviderType(req.Provider.ProviderType) {
	case deployment.Route53ProviderType:
		var hostedZoneId = req.Provider.HostedZoneID
		var recordName = req.Provider.RecordName
		var recordType = req.Provider.RecordType
		var setIdentifier = req.Provider.SetIdentifier
		p, err := deployment.NewRoute53ProviderSpec(hostedZoneId, recordName, recordType, setIdentifier)
		if err != nil {
			return nil, err
		}
		provider = p
	case deployment.InMemoryProviderType:
		p, err := deployment.NewInMemoryProviderSpec()
		if err != nil {
			return nil, err
		}
		provider = p
	default:
		return nil, fmt.Errorf("")
	}
	return provider, nil
}

func (c *Controller) newStepBehaviorWithPutDeploymentRequest(req *request.PutDeploymentRequest) (*deployment.StepBehaviorSpec, error) {
	var stepBehavior *deployment.StepBehaviorSpec
	switch deployment.StepBehaviorAlgorithm(req.Step.Algorithm) {
	case deployment.IncreaseBehaviorType:
		s, err := deployment.NewIncreaseStepBehaviorSpec(req.Step.Threshold)
		if err != nil {
			return nil, err
		}
		stepBehavior = s
	default:
		return nil, fmt.Errorf("")
	}

	return stepBehavior, nil
}

func (c *Controller) newRollbackBehaviorWithPutDeploymentRequest(req *request.PutDeploymentRequest) (*deployment.RollbackBehaviorSpec, error) {
	var rollbackBehavior *deployment.RollbackBehaviorSpec
	switch deployment.RollbackBehaviorAlgorithm(req.Rollback.Algorithm) {
	case deployment.RollbackDecreaseBehaviorType:
		r, err := deployment.NewDecreaseRollbackBehaviorSpec(req.Step.Threshold)
		if err != nil {
			return nil, err
		}
		rollbackBehavior = r
	case deployment.RollbackHistoryBehaviorType:
		r, err := deployment.NewHistoryRollbackBehaviorSpec()
		if err != nil {
			return nil, err
		}
		rollbackBehavior = r
	default:
		return nil, nil
	}

	return rollbackBehavior, nil
}

func (c *Controller) newMetricsWithPutDeploymentRequest(req *request.PutDeploymentRequest) ([]*deployment.MetricsSpec, error) {
	metrics := make([]*deployment.MetricsSpec, len(req.Metrics))
	for i, v := range req.Metrics {
		switch deployment.MetricsType(v.MetricType) {
		case deployment.InMemoryMetricsType:
			var percentage *float64
			var timeWindow *time.Duration
			if v.Target != nil {
				percentage = &v.Target.Percentage
				timeWindow = &v.Target.TimeWindow
			}
			m, err := deployment.NewInMemoryMetrics(v.Period, v.Condition, percentage, timeWindow)
			if err != nil {
				return nil, err
			}
			metrics[i] = m
		case deployment.CloudWatchMetricsType:
			var percentage *float64
			var timeWindow *time.Duration
			if v.Target != nil {
				percentage = &v.Target.Percentage
				timeWindow = &v.Target.TimeWindow
			}
			m, err := deployment.NewCloudWatchMetrics(v.Period, v.Condition, v.Query, *v.AllowNoData, percentage, timeWindow)
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

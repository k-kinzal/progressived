package response

import (
	"github.com/k-kinzal/progressived/pkg/progressived/persistence"
	v1 "github.com/k-kinzal/progressived/pkg/progressived/persistence/v1/deployment"
	"time"
)

type DeploymentProviderSpec struct {
	ProviderType  string `json:"type"`
	HostedZoneID  string `json:"hostedZoneId,omitempty"`
	RecordName    string `json:"recordName,omitempty"`
	RecordType    string `json:"recordType,omitempty"`
	SetIdentifier string `json:"setIdentifier,omitempty"`
}

type DeploymentStepBehaviorSpec struct {
	Algorithm string   `json:"algorithm"`
	Threshold *float64 `json:"threshold,omitempty"`
}

type DeploymentRollbackBehaviorSpec struct {
	RollbackBehaviorAlgorithm string   `json:"algorithm"`
	Threshold                 *float64 `json:"threshold,omitempty"`
}

type DeploymentMetricsTargetSpec struct {
	Percentage float64       `json:"percentage"`
	TimeWindow time.Duration `json:"timeWindow"`
}

type DeploymentMetricsSpec struct {
	MetricType  string                       `json:"name"`
	Period      time.Duration                `json:"period"`
	Condition   string                       `json:"condition"`
	Query       string                       `json:"query,omitempty"`
	AllowNoData *bool                        `json:"allowNoData,omitempty"`
	Target      *DeploymentMetricsTargetSpec `json:"target,omitempty"`
}

type DeploymentSpec struct {
	Interval time.Duration                   `json:"interval"`
	Provider *DeploymentProviderSpec         `json:"provider"`
	Step     *DeploymentStepBehaviorSpec     `json:"step"`
	Rollback *DeploymentRollbackBehaviorSpec `json:"rollback"`
	Metrics  []*DeploymentMetricsSpec        `json:"metrics,omitempty"`
}

type DeploymentScheduleState struct {
	Weight            float64   `json:"weight"`
	NextScheduledTime time.Time `json:"nextScheduledTime"`
}

type DeploymentRetryState struct {
	Reason            string    `json:"reason,omitempty"`
	Count             int       `json:"count"`
	NextScheduledTime time.Time `json:"nextScheduledTime"`
}

type DeploymentState struct {
	Revision  int                      `json:"revision"`
	Status    string                   `json:"status"`
	Weight    float64                  `json:"weight"`
	Schedule  *DeploymentScheduleState `json:"schedule,omitempty"`
	Retry     *DeploymentRetryState    `json:"retry,omitempty"`
	CreatedAt time.Time                `json:"createdAt"`
	UpdatedAt time.Time                `json:"updatedAt"`
}

type DeploymentHistory struct {
	Event      string                 `json:"event"`
	Revision   int                    `json:"revision"`
	Status     string                 `json:"status"`
	Weight     float64                `json:"weight"`
	CreatedAt  time.Time              `json:"createdAt"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type Deployment struct {
	Version string               `json:"version"`
	Name    string               `json:"name"`
	Spec    *DeploymentSpec      `json:"spec"`
	State   *DeploymentState     `json:"state"`
	History []*DeploymentHistory `json:"history"`
}

func newDeploymentWithEntityV1(entity *v1.Deployment) *Deployment {
	var provider *DeploymentProviderSpec
	switch persistence.ProviderType(entity.Spec.Provider.ProviderType) {
	case persistence.InMemoryProviderType:
		provider = &DeploymentProviderSpec{
			ProviderType: entity.Spec.Provider.ProviderType,
		}
	case persistence.Route53ProviderType:
		provider = &DeploymentProviderSpec{
			ProviderType:  entity.Spec.Provider.ProviderType,
			HostedZoneID:  entity.Spec.Provider.Route53.HostedZoneID,
			RecordName:    entity.Spec.Provider.Route53.RecordName,
			RecordType:    entity.Spec.Provider.Route53.RecordName,
			SetIdentifier: entity.Spec.Provider.Route53.SetIdentifier,
		}
	default:
		panic("")
	}

	var stepBehavior *DeploymentStepBehaviorSpec
	switch persistence.BehaviorAlgorithm(entity.Spec.Step.StepBehaviorAlgorithm) {
	case persistence.IncreaseBehaviorType:
		stepBehavior = &DeploymentStepBehaviorSpec{
			Algorithm: entity.Spec.Step.StepBehaviorAlgorithm,
			Threshold: &entity.Spec.Step.Increase.Threshold,
		}
	default:
		panic("")
	}

	var rollbackBehavior *DeploymentRollbackBehaviorSpec
	switch persistence.RollbackBehaviorAlgorithm(entity.Spec.Rollback.RollbackBehaviorAlgorithm) {
	case persistence.RollbackDecreaseBehaviorType:
		rollbackBehavior = &DeploymentRollbackBehaviorSpec{
			RollbackBehaviorAlgorithm: entity.Spec.Rollback.RollbackBehaviorAlgorithm,
			Threshold:                 &entity.Spec.Rollback.Decrease.Threshold,
		}
	case persistence.RollbackHistoryBehaviorType:
		rollbackBehavior = &DeploymentRollbackBehaviorSpec{
			RollbackBehaviorAlgorithm: entity.Spec.Rollback.RollbackBehaviorAlgorithm,
		}
	default:
		panic("")
	}

	var metrics []*DeploymentMetricsSpec = make([]*DeploymentMetricsSpec, len(entity.Spec.Metrics))
	for i, met := range entity.Spec.Metrics {
		switch persistence.MetricsType(met.MetricType) {
		case persistence.InMemoryMetricsType:
			metrics[i] = &DeploymentMetricsSpec{
				MetricType: met.MetricType,
				Period:     met.Period,
				Condition:  met.Condition,
				Target: &DeploymentMetricsTargetSpec{
					Percentage: met.Target.Percentage,
					TimeWindow: met.Target.TimeWindow,
				},
			}
		case persistence.InCloudWatchMetricsType:
			metrics[i] = &DeploymentMetricsSpec{
				MetricType:  met.MetricType,
				Period:      met.Period,
				Condition:   met.Condition,
				Query:       met.Query,
				AllowNoData: met.AllowNoData,
				Target: &DeploymentMetricsTargetSpec{
					Percentage: met.Target.Percentage,
					TimeWindow: met.Target.TimeWindow,
				},
			}
		}
	}

	spec := &DeploymentSpec{
		Interval: entity.Spec.Interval,
		Provider: provider,
		Step:     stepBehavior,
		Rollback: rollbackBehavior,
		Metrics:  metrics,
	}

	var schedule *DeploymentScheduleState
	if entity.State.Schedule != nil {
		schedule = &DeploymentScheduleState{
			Weight:            entity.State.Schedule.Weight,
			NextScheduledTime: entity.State.Schedule.NextScheduledTime,
		}
	}

	var retry *DeploymentRetryState
	if entity.State.Retry != nil {
		retry = &DeploymentRetryState{
			Reason:            entity.State.Retry.Reason,
			Count:             entity.State.Retry.Count,
			NextScheduledTime: entity.State.Retry.NextScheduledTime,
		}
	}

	state := &DeploymentState{
		Revision:  entity.State.Revision,
		Status:    entity.State.Status.String(),
		Weight:    entity.State.Weight,
		Schedule:  schedule,
		Retry:     retry,
		CreatedAt: entity.State.CreatedAt,
		UpdatedAt: entity.State.UpdatedAt,
	}

	var history []*DeploymentHistory = make([]*DeploymentHistory, len(entity.History))
	for i, hist := range entity.History {
		history[i] = &DeploymentHistory{
			Event:      hist.Event.String(),
			Revision:   hist.Revision,
			Status:     hist.Status.String(),
			Weight:     hist.Weight,
			CreatedAt:  hist.CreatedAt,
			Attributes: hist.Attributes,
		}
	}

	return &Deployment{
		Version: entity.Ver.String(),
		Name:    entity.Name,
		Spec:    spec,
		State:   state,
		History: history,
	}
}

func newDeploymentWithEntity(entity persistence.Deployment) *Deployment {
	switch entity.Version() {
	case "v1":
		e, ok := entity.(*v1.Deployment)
		if !ok {
			panic("")
		}
		return newDeploymentWithEntityV1(e)
	default:
		panic("")
	}
}

type PutDeploymentResponse struct {
	Deployment *Deployment `json:"deployment"`
}

func NewPutDeploymentResponseWith(entity persistence.Deployment) *PutDeploymentResponse {
	return &PutDeploymentResponse{
		Deployment: newDeploymentWithEntity(entity),
	}
}

type DescribeDeploymentResponse struct {
	Deployment *Deployment `json:"deployment"`
}

func NewDescribeDeploymentResponseWith(entity persistence.Deployment) *DescribeDeploymentResponse {
	return &DescribeDeploymentResponse{
		Deployment: newDeploymentWithEntity(entity),
	}
}

type ListDeploymentResponse struct {
	Deployments []*Deployment `json:"deployments"`
}

func NewListDeploymentResponseWith(entities []persistence.Deployment) *ListDeploymentResponse {
	var deployments []*Deployment = make([]*Deployment, len(entities))
	for i, entity := range entities {
		deployments[i] = newDeploymentWithEntity(entity)
	}

	return &ListDeploymentResponse{
		Deployments: deployments,
	}
}

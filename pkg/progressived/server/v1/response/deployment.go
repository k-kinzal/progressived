package response

import (
	v1 "github.com/k-kinzal/progressived/pkg/progressived/persistence/v1/deployment"
	"time"
)

type DeploymentProviderSpecBody struct {
	ProviderType  string `json:"type"`
	HostedZoneID  string `json:"hostedZoneId,omitempty"`
	RecordName    string `json:"recordName,omitempty"`
	RecordType    string `json:"recordType,omitempty"`
	SetIdentifier string `json:"setIdentifier,omitempty"`
}

type DeploymentStepBehaviorSpecBody struct {
	Algorithm string   `json:"algorithm"`
	Threshold *float64 `json:"threshold,omitempty"`
}

type DeploymentRollbackBehaviorSpecBody struct {
	RollbackBehaviorAlgorithm string   `json:"algorithm"`
	Threshold                 *float64 `json:"threshold,omitempty"`
}

type DeploymentMetricsTargetSpecBody struct {
	Percentage float64       `json:"percentage"`
	TimeWindow time.Duration `json:"timeWindow"`
}

type DeploymentMetricsSpecBody struct {
	MetricType  string                           `json:"name"`
	Period      time.Duration                    `json:"period"`
	Condition   string                           `json:"condition"`
	Query       string                           `json:"query,omitempty"`
	AllowNoData *bool                            `json:"allowNoData,omitempty"`
	Target      *DeploymentMetricsTargetSpecBody `json:"target,omitempty"`
}

type DeploymentSpecBody struct {
	Interval time.Duration                       `json:"interval"`
	Provider *DeploymentProviderSpecBody         `json:"provider"`
	Step     *DeploymentStepBehaviorSpecBody     `json:"step"`
	Rollback *DeploymentRollbackBehaviorSpecBody `json:"rollback"`
	Metrics  []*DeploymentMetricsSpecBody        `json:"metrics,omitempty"`
}

type DeploymentScheduleStateBody struct {
	Weight            float64   `json:"weight"`
	NextScheduledTime time.Time `json:"nextScheduledTime"`
}

type DeploymentRetryStateBody struct {
	Reason            string    `json:"reason,omitempty"`
	Count             int       `json:"count"`
	NextScheduledTime time.Time `json:"nextScheduledTime"`
}

type DeploymentStateBody struct {
	Revision  int                          `json:"revision"`
	Status    string                       `json:"status"`
	Weight    float64                      `json:"weight"`
	Schedule  *DeploymentScheduleStateBody `json:"schedule,omitempty"`
	Retry     *DeploymentRetryStateBody    `json:"retry,omitempty"`
	CreatedAt time.Time                    `json:"createdAt"`
	UpdatedAt time.Time                    `json:"updatedAt"`
}

type DeploymentHistoryBody struct {
	Event      string                 `json:"event"`
	Revision   int                    `json:"revision"`
	Status     string                 `json:"status"`
	Weight     float64                `json:"weight"`
	CreatedAt  time.Time              `json:"createdAt"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type DeploymentBody struct {
	Version string                   `json:"version"`
	Name    string                   `json:"name"`
	Spec    *DeploymentSpecBody      `json:"spec"`
	State   *DeploymentStateBody     `json:"state"`
	History []*DeploymentHistoryBody `json:"history"`
}

func newDeploymentBodyWithEntityV1(entity *v1.Deployment) *DeploymentBody {
	var provider *DeploymentProviderSpecBody
	switch entity.Provider().Type() {
	case v1.InMemoryProviderType:
		provider = &DeploymentProviderSpecBody{
			ProviderType: v1.InMemoryProviderType.String(),
		}
	case v1.Route53ProviderType:
		provider = &DeploymentProviderSpecBody{
			ProviderType:  v1.Route53ProviderType.String(),
			HostedZoneID:  entity.Provider().Route53Config().HostedZoneID(),
			RecordName:    entity.Provider().Route53Config().RecordName(),
			RecordType:    entity.Provider().Route53Config().RecordName(),
			SetIdentifier: entity.Provider().Route53Config().SetIdentifier(),
		}
	default:
		panic("")
	}

	var stepBehavior *DeploymentStepBehaviorSpecBody
	switch entity.StepBehavior().Algorithm() {
	case v1.IncreaseBehaviorType:
		stepBehavior = &DeploymentStepBehaviorSpecBody{
			Algorithm: v1.IncreaseBehaviorType.String(),
			Threshold: &entity.StepBehavior().IncreaseSpec().Threshold,
		}
	default:
		panic("")
	}

	var rollbackBehavior *DeploymentRollbackBehaviorSpecBody
	switch entity.RollbackBehavior().Algorithm() {
	case v1.RollbackHistoryBehaviorType:
		rollbackBehavior = &DeploymentRollbackBehaviorSpecBody{
			RollbackBehaviorAlgorithm: v1.RollbackHistoryBehaviorType.String(),
		}
	case v1.RollbackDecreaseBehaviorType:
		rollbackBehavior = &DeploymentRollbackBehaviorSpecBody{
			RollbackBehaviorAlgorithm: v1.RollbackDecreaseBehaviorType.String(),
			Threshold:                 &entity.RollbackBehavior().DecreaseSpec().Threshold,
		}
	default:
		panic("")
	}

	var metrics = make([]*DeploymentMetricsSpecBody, len(entity.Metrics()))
	for i, met := range entity.Metrics() {
		switch met.Type() {
		case v1.InMemoryMetricsType:
			metrics[i] = &DeploymentMetricsSpecBody{
				MetricType: v1.InMemoryMetricsType.String(),
				Period:     met.Period(),
				Condition:  met.Condition(),
				Target:     nil,
			}
			if target := met.Target(); target != nil {
				metrics[i].Target = &DeploymentMetricsTargetSpecBody{
					Percentage: target.Percentage(),
					TimeWindow: target.TimeWindow(),
				}
			}
		case v1.CloudWatchMetricsType:
			allowNoData := met.CloudWatchSpec().AllowNoData()
			metrics[i] = &DeploymentMetricsSpecBody{
				MetricType:  v1.CloudWatchMetricsType.String(),
				Period:      met.Period(),
				Condition:   met.Condition(),
				Query:       met.CloudWatchSpec().Query(),
				AllowNoData: &allowNoData,
				Target:      nil,
			}
			if target := met.Target(); target != nil {
				metrics[i].Target = &DeploymentMetricsTargetSpecBody{
					Percentage: target.Percentage(),
					TimeWindow: target.TimeWindow(),
				}
			}
		}
	}

	spec := &DeploymentSpecBody{
		Interval: entity.Interval(),
		Provider: provider,
		Step:     stepBehavior,
		Rollback: rollbackBehavior,
		Metrics:  metrics,
	}

	var schedule *DeploymentScheduleStateBody
	if s := entity.State().Schedule(); s != nil {
		schedule = &DeploymentScheduleStateBody{
			Weight:            s.Weight(),
			NextScheduledTime: s.NextScheduledTime(),
		}
	}

	var retry *DeploymentRetryStateBody
	if r := entity.State().Retry(); r != nil {
		retry = &DeploymentRetryStateBody{
			Reason:            r.Reason(),
			Count:             int(r.Count()),
			NextScheduledTime: r.NextScheduledTime(),
		}
	}

	state := &DeploymentStateBody{
		Revision:  int(entity.State().Revision()),
		Status:    entity.State().Status().String(),
		Weight:    entity.State().Weight(),
		Schedule:  schedule,
		Retry:     retry,
		CreatedAt: entity.State().CreatedAt(),
		UpdatedAt: entity.State().UpdatedAt(),
	}

	var history = make([]*DeploymentHistoryBody, len(entity.History()))
	for i, hist := range entity.History() {
		history[i] = &DeploymentHistoryBody{
			Event:      hist.Event().String(),
			Revision:   int(hist.Revision()),
			Status:     hist.Status().String(),
			Weight:     hist.Weight(),
			CreatedAt:  hist.CreatedAt(),
			Attributes: hist.Attributes(),
		}
	}

	return &DeploymentBody{
		Version: entity.Version().String(),
		Name:    entity.Name(),
		Spec:    spec,
		State:   state,
		History: history,
	}
}

type PutDeploymentResponse struct {
	Deployment *DeploymentBody `json:"deployment"`
}

func NewPutDeploymentResponse(entity *v1.Deployment) *PutDeploymentResponse {
	return &PutDeploymentResponse{
		Deployment: newDeploymentBodyWithEntityV1(entity),
	}
}

type DescribeDeploymentResponse struct {
	Deployment *DeploymentBody `json:"deployment"`
}

func NewDescribeDeploymentResponse(entity *v1.Deployment) *DescribeDeploymentResponse {
	return &DescribeDeploymentResponse{
		Deployment: newDeploymentBodyWithEntityV1(entity),
	}
}

type ListDeploymentResponse struct {
	Deployments []*DeploymentBody `json:"deployments"`
}

func NewListDeploymentResponse(entities []*v1.Deployment) *ListDeploymentResponse {
	var deployments []*DeploymentBody = make([]*DeploymentBody, len(entities))
	for i, entity := range entities {
		deployments[i] = newDeploymentBodyWithEntityV1(entity)
	}

	return &ListDeploymentResponse{
		Deployments: deployments,
	}
}

type ScheduleDeploymentResponse struct {
	Deployment *DeploymentBody `json:"deployment"`
}

func NewScheduleDeploymentResponse(entity *v1.Deployment) *ScheduleDeploymentResponse {
	return &ScheduleDeploymentResponse{
		Deployment: newDeploymentBodyWithEntityV1(entity),
	}
}

type PauseDeploymentResponse struct {
	Deployment *DeploymentBody `json:"deployment"`
}

func NewPauseDeploymentResponse(entity *v1.Deployment) *PauseDeploymentResponse {
	return &PauseDeploymentResponse{
		Deployment: newDeploymentBodyWithEntityV1(entity),
	}
}

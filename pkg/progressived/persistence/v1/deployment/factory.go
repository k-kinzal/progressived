package deployment

import (
	"github.com/k-kinzal/progressived/pkg/progressived/persistence"
	"time"
)

type Factory struct {
}

func (f *Factory) NewInMemoryProviderSpec() (persistence.Provider, error) {
	return &ProviderSpec{
		ProviderType: string(persistence.InMemoryProviderType),
		InMemory:     &InMemoryProviderSpec{},
	}, nil
}

func (f *Factory) NewRoute53ProviderSpec(
	hostedZoneID string,
	recordName string,
	recordType string,
	targetSetIdentifier string,
) (persistence.Provider, error) {
	return &ProviderSpec{
		ProviderType: string(persistence.Route53ProviderType),
		Route53: &Route53ProviderSpec{
			HostedZoneID:  hostedZoneID,
			RecordName:    recordName,
			RecordType:    recordType,
			SetIdentifier: targetSetIdentifier,
		},
	}, nil
}

func (f *Factory) NewIncreaseStepBehaviorSpec(threshold float64) (persistence.StepBehavior, error) {
	return &StepBehaviorSpec{
		StepBehaviorAlgorithm: string(persistence.IncreaseBehaviorType),
		Increase: &IncreaseStepBehaviorSpec{
			Threshold: threshold,
		},
	}, nil
}

func (f *Factory) NewDecreaseRollbackBehaviorSpec(threshold float64) (persistence.RollbackBehavior, error) {
	return &RollbackBehaviorSpec{
		RollbackBehaviorAlgorithm: string(persistence.RollbackDecreaseBehaviorType),
		Decrease: &DecreaseRollbackBehaviorSpec{
			Threshold: threshold,
		},
	}, nil
}

func (f *Factory) NewHistoryRollbackBehaviorSpec() (persistence.RollbackBehavior, error) {
	return &RollbackBehaviorSpec{
		RollbackBehaviorAlgorithm: string(persistence.RollbackHistoryBehaviorType),
		History:                   &HistoryRollbackBehaviorSpec{},
	}, nil
}

func (f *Factory) NewInMemoryMetrics(
	period time.Duration,
	condition string,
	percentage float64,
	timeWindow time.Duration,
) (persistence.Metrics, error) {
	return &MetricsSpec{
		MetricType:  string(persistence.InMemoryMetricsType),
		Period:      period,
		Condition:   condition,
		Query:       "",
		AllowNoData: nil,
		Target: &MetricsTargetSpec{
			Percentage: percentage,
			TimeWindow: timeWindow,
		},
	}, nil
}

func (f *Factory) NewCloudWatchMetrics(
	period time.Duration,
	condition string,
	query string,
	allowNoData bool,
	percentage float64,
	timeWindow time.Duration,
) (persistence.Metrics, error) {
	return &MetricsSpec{
		MetricType:  string(persistence.InCloudWatchMetricsType),
		Period:      period,
		Condition:   condition,
		Query:       query,
		AllowNoData: &allowNoData,
		Target: &MetricsTargetSpec{
			Percentage: percentage,
			TimeWindow: timeWindow,
		},
	}, nil
}

func (f *Factory) New(
	name string,
	interval time.Duration,
	provider persistence.Provider,
	step persistence.StepBehavior,
	rollback persistence.RollbackBehavior,
	metrics []persistence.Metrics,
) (persistence.Deployment, error) {
	spec := Spec{
		Interval: interval,
		Provider: provider.(*ProviderSpec),
		Step:     step.(*StepBehaviorSpec),
		Rollback: nil,
		Metrics:  make([]*MetricsSpec, len(metrics)),
	}

	if rollback == nil {
		defaultRollback, err := f.NewHistoryRollbackBehaviorSpec()
		if err != nil {
			return nil, err
		}
		spec.Rollback = defaultRollback.(*RollbackBehaviorSpec)
	} else {
		spec.Rollback = rollback.(*RollbackBehaviorSpec)
	}

	for i, v := range metrics {
		spec.Metrics[i] = v.(*MetricsSpec)
	}

	entity := &Deployment{
		Ver:  V1,
		Name: name,
		Spec: spec,
		State: State{
			Revision:  1,
			Status:    "ready",
			Weight:    -1,
			Schedule:  nil,
			Retry:     nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	return entity, nil
}

func NewFactory() *Factory {
	return &Factory{}
}

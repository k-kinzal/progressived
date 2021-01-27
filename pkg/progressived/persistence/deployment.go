package persistence

import (
	"time"
)

type ProviderType string

func (p *ProviderType) String() string {
	return string(*p)
}

const (
	InMemoryProviderType ProviderType = "inmemory"
	Route53ProviderType  ProviderType = "route53"
)

type Provider interface {
	Type() ProviderType
}

type BehaviorAlgorithm string

func (t *BehaviorAlgorithm) String() string {
	return string(*t)
}

const (
	IncreaseBehaviorType BehaviorAlgorithm = "increase"
)

type StepBehavior interface {
	Algorithm() BehaviorAlgorithm
}

type RollbackBehaviorAlgorithm string

func (t *RollbackBehaviorAlgorithm) String() string {
	return string(*t)
}

const (
	RollbackDecreaseBehaviorType RollbackBehaviorAlgorithm = "decrease"
	RollbackHistoryBehaviorType  RollbackBehaviorAlgorithm = "history"
)

type RollbackBehavior interface {
	Algorithm() RollbackBehaviorAlgorithm
}

type MetricsType string

func (t *MetricsType) String() string {
	return string(*t)
}

const (
	InMemoryMetricsType     MetricsType = "inmemory"
	InCloudWatchMetricsType MetricsType = "cloudwatch"
)

type Metrics interface {
	Type() MetricsType
}

type StateStatus string

func (s *StateStatus) String() string {
	return string(*s)
}

const (
	ReadyStateStatus StateStatus = "ready"
)

type Deployment interface {
	// TODO: add behavior
	Version() string
	Status() StateStatus
	ProviderType() ProviderType
	Update(interval time.Duration, provider Provider, step StepBehavior, rollback RollbackBehavior, metrics []Metrics) (Deployment, error)
}

type DeploymentFactory interface {
	NewInMemoryProviderSpec() (Provider, error)
	NewRoute53ProviderSpec(hostedZoneID string, recordName string, recordType string, setIdentifier string) (Provider, error)
	NewIncreaseStepBehaviorSpec(threshold float64) (StepBehavior, error)
	NewDecreaseRollbackBehaviorSpec(threshold float64) (RollbackBehavior, error)
	NewHistoryRollbackBehaviorSpec() (RollbackBehavior, error)
	NewInMemoryMetrics(period time.Duration, condition string, percentage float64, timeWindow time.Duration) (Metrics, error)
	NewCloudWatchMetrics(period time.Duration, condition string, query string, allowNoData bool, percentage float64, timeWindow time.Duration) (Metrics, error)
	New(name string, interval time.Duration, provider Provider, step StepBehavior, rollback RollbackBehavior, metrics []Metrics) (Deployment, error)
}

type DeploymentFilterFunc func(Deployment) bool

type Deployments interface {
	Put(element Deployment) error
	Get(name string) (Deployment, error)
	Seq() ([]Deployment, error)
	Filter(filter DeploymentFilterFunc) (DeploymentsReadOnly, error)
	Changes() <-chan Deployment
}

type DeploymentsReadOnly interface {
	Get(name string) (Deployment, error)
	Seq() ([]Deployment, error)
	Filter(filter DeploymentFilterFunc) (DeploymentsReadOnly, error)
	Changes() <-chan Deployment
}

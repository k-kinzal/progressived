package deployment

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/k-kinzal/progressived/pkg/progressived/persistence"
	"time"
)

// version: v1
// name: foo-deployment
// spec:
//   interval: 5m
//   provider:
//     type: route53
//     hostedZoneId: xxx
//     recordName: example.com
//     type: AAAA
//     setIdentifier: xxx
//   step:
//     algorithm: step
//     threshold: 10
//   rollback:
//     algorithm: history
//   metrics:
//   - type: inmemory
//     period: 1m
//     condition: x > 10
//   - type: cloudwatch
//     period: 1m
//     query: '{"Id":"q1","MetricStat":{"Metric":{"Namespace":"JawsCLI","MetricName":"TestMetric","Dimensions":[{"Name":"TestKey","Value":"TestValue"}]},"Period":60,"Stat":"Average"}}'
//     allowNoData: false
//     condition: x > 10
//     target:
//       percentage: 99.9%
//       timeWindow: 50m
// state:
//   revision: 7
//   status: progress
//   weight: 10
//   schedule:
//     weight: 20
//     nextScheduledTime: 2021-01-01T00:00:00+09:00Z
//   retry:
//     reason: none
//     count: 0
//     nextScheduledTime: nil
//   createdAt: 2021-01-01T00:00:00+09:00Z
//   updatedAt: 2021-01-01T00:00:00+09:00Z
// history:
// - {"event":"READY","status":"ready","weight":0,"createdAt":"2021-01-01T00:00:00+09:00Z"}
// - {"event":"ADD_STATE","status":"progress","weight":10,"createdAt":"2021-01-01T00:00:00+09:00Z"}
// - {"event":"ROLLBACK","status":"rollback","revision":1,"weight":10,"reason":"invalid condition `x(100) > 10`","createdAt":"2021-01-01T00:00:00+09:00Z"}
// - {"event":"ROLLBACK_COMPLETE","status":"rollback complete","revision":0,"weight":0,"reason":"invalid condition `x(100) > 10`","createdAt":"2021-01-01T00:00:00+09:00Z"
// - {"event":"RE_READY","status":"re-ready","weight":0,"createdAt":"2021-01-01T00:00:00+09:00Z"}
// - {"event":"ADD_STATE","status":"progress","weight":10,"createdAt":"2021-01-01T00:00:00+09:00Z"}

type VersionV1 string

const V1 VersionV1 = "v1"

func (v *VersionV1) MarshalJSON() ([]byte, error) {
	return json.Marshal(V1)
}
func (v *VersionV1) UnmarshalJSON(data []byte) error {
	if string(data) != "v1" {
		return errors.New("")
	}
	*v = V1
	return nil
}
func (v *VersionV1) String() string {
	return "v1"
}

type InMemoryProviderSpec struct {
}

type Route53ProviderSpec struct {
	HostedZoneID  string `json:"hostedZoneId" validate:"required,fqdn"`
	RecordName    string `json:"recordName" validate:"required,fqdn"`
	RecordType    string `json:"recordType,omitempty" validate:"oneof=A AAAA CNAME"`
	SetIdentifier string `json:"setIdentifier,omitempty"`
}

type ProviderSpec struct {
	ProviderType string `json:"type" validate:"required,oneof=inmemory route53"`

	// inmemory
	InMemory *InMemoryProviderSpec `json:"inmemory,omitempty" validate:"required_if=ProviderType inmemory"`

	// route53
	Route53 *Route53ProviderSpec `json:"route53,omitempty" validate:"required_if=ProviderType route53"`
}

func (p *ProviderSpec) Type() persistence.ProviderType {
	return persistence.ProviderType(p.ProviderType)
}

type IncreaseStepBehaviorSpec struct {
	Threshold float64 `json:"threshold" validate:"required,min=0,max=100"`
}

type StepBehaviorSpec struct {
	StepBehaviorAlgorithm string `json:"algorithm" validate:"required,oneof=increase"`

	// increase
	Increase *IncreaseStepBehaviorSpec `json:"increase,omitempty" validate:"required_if=StepBehaviorAlgorithm increase"`
}

func (b *StepBehaviorSpec) Algorithm() persistence.BehaviorAlgorithm {
	return persistence.BehaviorAlgorithm(b.StepBehaviorAlgorithm)
}

type DecreaseRollbackBehaviorSpec struct {
	Threshold float64 `json:"threshold" validate:"required,min=0,max=100"`
}

type HistoryRollbackBehaviorSpec struct {
}

type RollbackBehaviorSpec struct {
	RollbackBehaviorAlgorithm string `json:"algorithm" validate:"required,oneof=decrease history"`

	// decrease
	Decrease *DecreaseRollbackBehaviorSpec `json:"decrease,omitempty" validate:"required_if=RollbackBehaviorAlgorithm decrease"`

	// history
	History *HistoryRollbackBehaviorSpec `json:"history,omitempty" validate:"required_if=RollbackBehaviorAlgorithm history"`
}

func (b *RollbackBehaviorSpec) Algorithm() persistence.RollbackBehaviorAlgorithm {
	return persistence.RollbackBehaviorAlgorithm(b.RollbackBehaviorAlgorithm)
}

type MetricsTargetSpec struct {
	Percentage float64       `json:"percentage" validate:"required,min=0,max=1"`
	TimeWindow time.Duration `json:"timeWindow" validate:"required"`
}

type MetricsSpec struct {
	MetricType  string             `json:"name" validate:"required,oneof=inmemory cloudwatch"`
	Period      time.Duration      `json:"period" validate:"required"`
	Condition   string             `json:"condition" validate:"required"`
	Query       string             `json:"query,omitempty" validate:"required_if=MetricType cloudwatch"`
	AllowNoData *bool              `json:"allowNoData,omitempty" validate:"required_if=MetricType cloudwatch"`
	Target      *MetricsTargetSpec `json:"target,omitempty"`
}

func (m *MetricsSpec) Type() persistence.MetricsType {
	return persistence.MetricsType(m.MetricType)
}

type Spec struct {
	Interval time.Duration         `json:"interval" validate:"required"`
	Provider *ProviderSpec         `json:"provider" validate:"required"`
	Step     *StepBehaviorSpec     `json:"step,omitempty" validate:"required"`
	Rollback *RollbackBehaviorSpec `json:"rollback,omitempty" validate:"required"`
	Metrics  []*MetricsSpec        `json:"metrics,omitempty"`
}

type ScheduleState struct {
	Weight            float64   `json:"weight" validate:"required"`
	NextScheduledTime time.Time `json:"nextScheduledTime" validate:"required"`
}

type RetryState struct {
	Reason            string    `json:"reason,omitempty"`
	Count             int       `json:"count" validate:"required"`
	NextScheduledTime time.Time `json:"nextScheduledTime" validate:"required"`
}

type State struct {
	Revision int            `json:"revision" validate:"required"`
	Status   StateStatus    `json:"status" validate:"required"`
	Weight   float64        `json:"weight"`
	Schedule *ScheduleState `json:"schedule,omitempty"`
	Retry    *RetryState    `json:"retry,omitempty"`

	CreatedAt time.Time `json:"createdAt" validate:"required"`
	UpdatedAt time.Time `json:"updatedAt" validate:"required"`
}

type EventType string

func (e *EventType) String() string {
	return string(*e)
}

// - {"event":"READY","status":"ready","weight":0,"createdAt":"2021-01-01T00:00:00+09:00Z"}
// - {"event":"ADD_STATE","status":"progress","weight":10,"createdAt":"2021-01-01T00:00:00+09:00Z"}
// - {"event":"ROLLBACK","status":"rollback","revision":1,"weight":10,"reason":"invalid condition `x(100) > 10`","createdAt":"2021-01-01T00:00:00+09:00Z"}
// - {"event":"ROLLBACK_COMPLETE","status":"rollback complete","revision":0,"weight":0,"reason":"invalid condition `x(100) > 10`","createdAt":"2021-01-01T00:00:00+09:00Z"
// - {"event":"RE_READY","status":"re-ready","weight":0,"createdAt":"2021-01-01T00:00:00+09:00Z"}
// - {"event":"ADD_STATE","status":"progress","weight":10,"createdAt":"2021-01-01T00:00:00+09:00Z"}
type History struct {
	Event      EventType              `json:"event" validate:"required"`
	Revision   int                    `json:"revision" validate:"required"`
	Status     StateStatus            `json:"status" validate:"required"`
	Weight     float64                `json:"weight" validate:"required"`
	CreatedAt  time.Time              `json:"createdAt" validate:"required"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

type Deployment struct {
	Ver  VersionV1 `json:"version" validate:"required,eq=v1"`
	Name string    `json:"name" validate:"required"`

	Spec    Spec       `json:"spec"`
	State   State      `json:"state"`
	History []*History `json:"history"`
}

func (e *Deployment) Version() string {
	return e.Ver.String()
}

func (e *Deployment) Clone() *Deployment {
	var buf bytes.Buffer
	var data *Deployment
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(e); err != nil {
		panic(err)
	}
	dec := gob.NewDecoder(&buf)
	if err := dec.Decode(&data); err != nil {
		panic(err)
	}
	return data
}

func (e *Deployment) Update(
	interval time.Duration,
	provider persistence.Provider,
	step persistence.StepBehavior,
	rollback persistence.RollbackBehavior,
	metrics []persistence.Metrics) (persistence.Deployment, error) {
	entity := e.Clone()
	entity.Spec.Interval = interval
	entity.Spec.Provider = provider.(*ProviderSpec)
	entity.Spec.Step = step.(*StepBehaviorSpec)
	entity.Spec.Rollback = rollback.(*RollbackBehaviorSpec)
	entity.Spec.Metrics = make([]*MetricsSpec, len(metrics))
	for i, v := range metrics {
		entity.Spec.Metrics[i] = v.(*MetricsSpec)
	}
	entity.State = State{
		Revision:  entity.State.Revision + 1,
		Status:    ReadyStateStatus,
		Weight:    0,
		Schedule:  nil,
		Retry:     nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	entity.History = append(entity.History)

	validate := validator.New()
	if err := validate.Struct(entity); err != nil {
		return nil, err
	}

	return entity, nil
}

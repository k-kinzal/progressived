package deployment

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/k-kinzal/progressived/pkg/algorithm"
	"time"
)

// Version

type version string

const VersionV1 version = "v1"

func (v *version) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}
func (v *version) UnmarshalJSON(data []byte) error {
	switch version(data) {
	case VersionV1:
		*v = VersionV1
	default:
		return errors.New("")
	}
	return nil
}
func (v version) String() string {
	return string(v)
}

// Provider

type ProviderType string

func (p ProviderType) String() string {
	return string(p)
}

const (
	InMemoryProviderType ProviderType = "inmemory"
	Route53ProviderType  ProviderType = "route53"
)

type InMemoryProviderSpec struct {
}

type Route53ProviderSpec struct {
	hostedZoneID  string `json:"hostedZoneId" validate:"required,fqdn"`
	recordName    string `json:"recordName" validate:"required,fqdn"`
	recordType    string `json:"recordType,omitempty" validate:"oneof=A AAAA CNAME"`
	setIdentifier string `json:"setIdentifier,omitempty"`
}

func (s *Route53ProviderSpec) HostedZoneID() string  { return s.hostedZoneID }
func (s *Route53ProviderSpec) RecordName() string    { return s.recordName }
func (s *Route53ProviderSpec) RecordType() string    { return s.recordType }
func (s *Route53ProviderSpec) SetIdentifier() string { return s.setIdentifier }

type ProviderSpec struct {
	providerType ProviderType `json:"type"`

	// inmemory
	inMemory *InMemoryProviderSpec `json:"inmemory,omitempty" validate:"required_if=ProviderType inmemory"`

	// route53
	route53 *Route53ProviderSpec `json:"route53,omitempty" validate:"required_if=ProviderType route53"`
}

func (s *ProviderSpec) Type() ProviderType                    { return s.providerType }
func (s *ProviderSpec) InMemoryConfig() *InMemoryProviderSpec { return s.inMemory }
func (s *ProviderSpec) Route53Config() *Route53ProviderSpec   { return s.route53 }

func NewInMemoryProviderSpec() (*ProviderSpec, error) {
	spec := &ProviderSpec{
		providerType: InMemoryProviderType,
		inMemory:     &InMemoryProviderSpec{},
	}

	return spec, nil
}

func NewRoute53ProviderSpec(
	hostedZoneID string,
	recordName string,
	recordType string,
	targetSetIdentifier string) (*ProviderSpec, error) {
	spec := &ProviderSpec{
		providerType: Route53ProviderType,
		route53: &Route53ProviderSpec{
			hostedZoneID:  hostedZoneID,
			recordName:    recordName,
			recordType:    recordType,
			setIdentifier: targetSetIdentifier,
		},
	}

	validate := validator.New()
	if err := validate.Struct(spec); err != nil {
		return nil, err
	}

	return spec, nil
}

// Step Behavior

type StepBehaviorAlgorithm string

func (t StepBehaviorAlgorithm) String() string {
	return string(t)
}

const (
	IncreaseBehaviorType StepBehaviorAlgorithm = "increase"
)

type IncreaseStepBehaviorSpec struct {
	Threshold float64 `json:"threshold" validate:"min=0,max=100"`
}

type StepBehaviorSpec struct {
	algorithm StepBehaviorAlgorithm `json:"algorithm"`

	// increase
	increase *IncreaseStepBehaviorSpec `json:"increase,omitempty" validate:"required_if=StepBehaviorAlgorithm increase"`
}

func (s *StepBehaviorSpec) Algorithm() StepBehaviorAlgorithm        { return s.algorithm }
func (s *StepBehaviorSpec) IncreaseSpec() *IncreaseStepBehaviorSpec { return s.increase }

func (s *StepBehaviorSpec) nextStep(weight float64) float64 {
	var nextWeight float64
	switch s.algorithm {
	case IncreaseBehaviorType:
		nextWeight = algorithm.NewIncretion(s.increase.Threshold).Next(weight)
	}
	return nextWeight
}

func NewIncreaseStepBehaviorSpec(threshold float64) (*StepBehaviorSpec, error) {
	spec := &StepBehaviorSpec{
		algorithm: IncreaseBehaviorType,
		increase: &IncreaseStepBehaviorSpec{
			Threshold: threshold,
		},
	}

	validate := validator.New()
	if err := validate.Struct(spec); err != nil {
		return nil, err
	}

	return spec, nil
}

// Rollback Behavior

type RollbackBehaviorAlgorithm string

func (t RollbackBehaviorAlgorithm) String() string {
	return string(t)
}

const (
	RollbackHistoryBehaviorType  RollbackBehaviorAlgorithm = "history"
	RollbackDecreaseBehaviorType RollbackBehaviorAlgorithm = "decrease"
)

type HistoryRollbackBehaviorSpec struct {
}

type DecreaseRollbackBehaviorSpec struct {
	Threshold float64 `json:"threshold" validate:"min=0,max=100"`
}

type RollbackBehaviorSpec struct {
	algorithm RollbackBehaviorAlgorithm `json:"algorithm"`

	// history
	history *HistoryRollbackBehaviorSpec `json:"history,omitempty" validate:"required_if=RollbackBehaviorAlgorithm history"`

	// decrease
	decrease *DecreaseRollbackBehaviorSpec `json:"decrease,omitempty" validate:"required_if=RollbackBehaviorAlgorithm decrease"`
}

func (s *RollbackBehaviorSpec) Algorithm() RollbackBehaviorAlgorithm        { return s.algorithm }
func (s *RollbackBehaviorSpec) HistorySpec() *HistoryRollbackBehaviorSpec   { return s.history }
func (s *RollbackBehaviorSpec) DecreaseSpec() *DecreaseRollbackBehaviorSpec { return s.decrease }

func NewHistoryRollbackBehaviorSpec() (*RollbackBehaviorSpec, error) {
	spec := &RollbackBehaviorSpec{
		algorithm: RollbackHistoryBehaviorType,
		history:   &HistoryRollbackBehaviorSpec{},
	}
	return spec, nil
}

func NewDecreaseRollbackBehaviorSpec(threshold float64) (*RollbackBehaviorSpec, error) {
	spec := &RollbackBehaviorSpec{
		algorithm: RollbackDecreaseBehaviorType,
		decrease: &DecreaseRollbackBehaviorSpec{
			Threshold: threshold,
		},
	}

	validate := validator.New()
	if err := validate.Struct(spec); err != nil {
		return nil, err
	}

	return spec, nil
}

// Metrics

type MetricsTargetSpec struct {
	percentage float64       `json:"percentage" validate:"min=0,max=1"`
	timeWindow time.Duration `json:"timeWindow" validate:"required"`
}

type MetricsType string

func (s *MetricsTargetSpec) Percentage() float64       { return s.percentage }
func (s *MetricsTargetSpec) TimeWindow() time.Duration { return s.timeWindow }

func (t MetricsType) String() string {
	return string(t)
}

const (
	InMemoryMetricsType   MetricsType = "inmemory"
	CloudWatchMetricsType MetricsType = "cloudwatch"
)

type InMemoryMetricsSpec struct {
	arr []float64 `json:"arr" validate:"required"`
}

type CloudWatchMetricsSpec struct {
	query       string `json:"query" validate:"required"`
	allowNoData bool   `json:"allowNoData"`
}

func (s *CloudWatchMetricsSpec) Query() string     { return s.query }
func (s *CloudWatchMetricsSpec) AllowNoData() bool { return s.allowNoData }

type MetricsSpec struct {
	metricType MetricsType        `json:"name"`
	period     time.Duration      `json:"period" validate:"required"`
	condition  string             `json:"condition" validate:"required"`
	target     *MetricsTargetSpec `json:"target,omitempty"`

	// inmemory
	inmemory *InMemoryMetricsSpec `json:"inmemory,omitempty" validate:"require_if=metricType inmemory"`

	// cloudwatch
	cloudwatch *CloudWatchMetricsSpec `json:"cloudwatch,omitempty" validate:"require_if=metricType cloudwatch"`
}

func (s *MetricsSpec) Type() MetricsType                      { return s.metricType }
func (s *MetricsSpec) Period() time.Duration                  { return s.period }
func (s *MetricsSpec) Condition() string                      { return s.condition }
func (s *MetricsSpec) Target() *MetricsTargetSpec             { return s.target }
func (s *MetricsSpec) InMemorySpec() *InMemoryMetricsSpec     { return s.inmemory }
func (s *MetricsSpec) CloudWatchSpec() *CloudWatchMetricsSpec { return s.cloudwatch }

func NewInMemoryMetrics(
	period time.Duration,
	condition string,
	percentage *float64,
	timeWindow *time.Duration) (*MetricsSpec, error) {
	var target *MetricsTargetSpec
	if percentage != nil && timeWindow != nil {
		target = &MetricsTargetSpec{
			percentage: *percentage,
			timeWindow: *timeWindow,
		}
	}

	spec := &MetricsSpec{
		metricType: InMemoryMetricsType,
		period:     period,
		condition:  condition,
		target:     target,
		inmemory: &InMemoryMetricsSpec{
			arr: []float64{1, 2, 3, 4},
		},
	}
	return spec, nil
}

func NewCloudWatchMetrics(
	period time.Duration,
	condition string,
	query string,
	allowNoData bool,
	percentage *float64,
	timeWindow *time.Duration) (*MetricsSpec, error) {
	var target *MetricsTargetSpec
	if percentage != nil && timeWindow != nil {
		target = &MetricsTargetSpec{
			percentage: *percentage,
			timeWindow: *timeWindow,
		}
	}

	spec := &MetricsSpec{
		metricType: CloudWatchMetricsType,
		period:     period,
		condition:  condition,
		target:     target,
		cloudwatch: &CloudWatchMetricsSpec{
			query:       query,
			allowNoData: allowNoData,
		},
	}
	return spec, nil
}

// Spec

type Spec struct {
	interval time.Duration         `json:"interval" validate:"required"`
	provider *ProviderSpec         `json:"provider" validate:"required"`
	step     *StepBehaviorSpec     `json:"step" validate:"required"`
	rollback *RollbackBehaviorSpec `json:"rollback" validate:"required"`
	metrics  []*MetricsSpec        `json:"metrics,omitempty"`
}

// State

type ScheduleState struct {
	weight            float64   `json:"weight" validate:"required"`
	nextScheduledTime time.Time `json:"nextScheduledTime" validate:"required"`
}

func (s *ScheduleState) Weight() float64              { return s.weight }
func (s *ScheduleState) NextScheduledTime() time.Time { return s.nextScheduledTime }

type RetryState struct {
	reason            string    `json:"reason,omitempty"`
	count             uint      `json:"count"`
	nextScheduledTime time.Time `json:"nextScheduledTime" validate:"required"`
}

func (s *RetryState) Reason() string               { return s.reason }
func (s *RetryState) Count() uint                  { return s.count }
func (s *RetryState) NextScheduledTime() time.Time { return s.nextScheduledTime }

type StateStatus string

func (s StateStatus) String() string {
	return string(s)
}

const (
	ReadyStateStatus             StateStatus = "ready"
	ProgressStateStatus          StateStatus = "progress"
	CompletedStateStatus         StateStatus = "completed"
	RollbackStateStatus          StateStatus = "rollback"
	RollbackCompletedStateStatus StateStatus = "rollback_completed"
	PauseStateStatus             StateStatus = "pause"
)

type State struct {
	revision uint           `json:"revision" validate:"required,min=0"`
	status   StateStatus    `json:"status" validate:"required"`
	weight   float64        `json:"weight"`
	schedule *ScheduleState `json:"schedule,omitempty"`
	retry    *RetryState    `json:"retry,omitempty"`

	createdAt time.Time `json:"createdAt" validate:"required"`
	updatedAt time.Time `json:"updatedAt" validate:"required"`
}

func (s *State) Revision() uint           { return s.revision }
func (s *State) Status() StateStatus      { return s.status }
func (s *State) Weight() float64          { return s.weight }
func (s *State) Schedule() *ScheduleState { return s.schedule }
func (s *State) Retry() *RetryState       { return s.retry }
func (s *State) CreatedAt() time.Time     { return s.createdAt }
func (s *State) UpdatedAt() time.Time     { return s.updatedAt }

// History

type EventType string

func (e EventType) String() string {
	return string(e)
}

const (
	CreateEventType         EventType = "CREATE DEPLOYMENT"
	UpdateSpecEventType     EventType = "UPDATE SPEC"
	SchedulingSpecEventType EventType = "SCHEDULING"
	PauseSpecEventType      EventType = "PAUSE"
)

type History struct {
	event      EventType              `json:"event"`
	revision   uint                   `json:"revision"`
	status     StateStatus            `json:"status"`
	weight     float64                `json:"weight"`
	createdAt  time.Time              `json:"createdAt"`
	attributes map[string]interface{} `json:"attributes,omitempty"`
}

func (h *History) Event() EventType                   { return h.event }
func (h *History) Revision() uint                     { return h.revision }
func (h *History) Status() StateStatus                { return h.status }
func (h *History) Weight() float64                    { return h.weight }
func (h *History) CreatedAt() time.Time               { return h.createdAt }
func (h *History) Attributes() map[string]interface{} { return h.attributes }

// Deployment

type Deployment struct {
	version version `json:"version"`
	name    string  `json:"name"`

	spec    *Spec      `json:"spec" validate:"required"`
	state   *State     `json:"state" validate:"required"`
	history []*History `json:"history" validate:"required"`
}

func (e *Deployment) Version() version                        { return e.version }
func (e *Deployment) Name() string                            { return e.name }
func (e *Deployment) Interval() time.Duration                 { return e.spec.interval }
func (e *Deployment) Provider() *ProviderSpec                 { return e.spec.provider }
func (e *Deployment) StepBehavior() *StepBehaviorSpec         { return e.spec.step }
func (e *Deployment) RollbackBehavior() *RollbackBehaviorSpec { return e.spec.rollback }
func (e *Deployment) Metrics() []*MetricsSpec                 { return e.spec.metrics }
func (e *Deployment) State() *State                           { return e.state }
func (e *Deployment) History() []*History                     { return e.history }

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
	provider *ProviderSpec,
	step *StepBehaviorSpec,
	rollback *RollbackBehaviorSpec,
	metrics []*MetricsSpec) (*Deployment, error) {
	now := time.Now()

	spec := &Spec{
		interval: interval,
		provider: provider,
		step:     step,
		rollback: rollback,
		metrics:  metrics,
	}

	if spec.rollback == nil {
		defaultRollback, err := NewHistoryRollbackBehaviorSpec()
		if err != nil {
			return nil, err
		}
		spec.rollback = defaultRollback
	}

	var previousSpec map[string]interface{}
	out, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(out, &previousSpec); err != nil {
		return nil, err
	}

	entity := &Deployment{
		version: e.version,
		name:    e.name,
		spec:    spec,
		state: &State{
			revision: uint(len(e.history) + 1),
			status:   ReadyStateStatus,
			weight:   0,
			schedule: &ScheduleState{
				weight:            spec.step.nextStep(0),
				nextScheduledTime: now.Add(spec.interval),
			},
			retry:     nil,
			createdAt: e.state.createdAt,
			updatedAt: now,
		},
		history: append(e.history, &History{
			event:     UpdateSpecEventType,
			revision:  uint(len(e.history) + 1),
			status:    ReadyStateStatus,
			weight:    0,
			createdAt: now,
			attributes: map[string]interface{}{
				"previousSpec": previousSpec,
			},
		}),
	}

	validate := validator.New()
	if err := validate.Struct(entity); err != nil {
		return nil, err
	}

	return entity, nil
}

func (e *Deployment) Scheduling(nextScheduledTime *time.Time) (*Deployment, error) {
	if nextScheduledTime == nil {
		t := time.Now().Add(e.spec.interval)
		nextScheduledTime = &t
	}

	entity := e.Clone()

	switch e.State().Status() {
	case ReadyStateStatus, PauseStateStatus:
		entity.state = &State{
			revision: uint(len(e.history) + 1),
			status:   ProgressStateStatus,
			weight:   e.state.weight,
			schedule: &ScheduleState{
				weight:            e.spec.step.nextStep(e.state.weight),
				nextScheduledTime: *nextScheduledTime,
			},
			retry:     e.state.retry,
			createdAt: e.state.createdAt,
			updatedAt: time.Now(),
		}
		entity.history = append(e.history, &History{
			event:     SchedulingSpecEventType,
			revision:  entity.state.revision,
			status:    entity.state.status,
			weight:    entity.state.weight,
			createdAt: entity.state.createdAt,
			attributes: map[string]interface{}{
				"nextScheduledWeight": entity.state.schedule.weight,
				"nextScheduledTime":   entity.state.schedule.nextScheduledTime,
			},
		})
	default:
		return nil, fmt.Errorf("can only be executed when %s or %s", ReadyStateStatus, ReadyStateStatus)
	}

	return entity, nil
}

func (e *Deployment) Pause() (*Deployment, error) {
	entity := e.Clone()

	switch e.State().Status() {
	case ProgressStateStatus, RollbackStateStatus:
		entity.state = &State{
			revision:  uint(len(e.history) + 1),
			status:    PauseStateStatus,
			weight:    e.state.weight,
			schedule:  nil,
			retry:     nil,
			createdAt: e.state.createdAt,
			updatedAt: time.Now(),
		}
		entity.history = append(e.history, &History{
			event:      PauseSpecEventType,
			revision:   entity.state.revision,
			status:     entity.state.status,
			weight:     entity.state.weight,
			createdAt:  entity.state.createdAt,
			attributes: map[string]interface{}{},
		})
	default:
		return nil, fmt.Errorf("can only be executed when %s or %s", ProgressStateStatus, RollbackStateStatus)
	}

	return entity, nil
}

func New(
	name string,
	interval time.Duration,
	provider *ProviderSpec,
	step *StepBehaviorSpec,
	rollback *RollbackBehaviorSpec,
	metrics []*MetricsSpec) (*Deployment, error) {
	now := time.Now()

	spec := &Spec{
		interval: interval,
		provider: provider,
		step:     step,
		rollback: rollback,
		metrics:  metrics,
	}

	if spec.rollback == nil {
		defaultRollback, err := NewHistoryRollbackBehaviorSpec()
		if err != nil {
			return nil, err
		}
		spec.rollback = defaultRollback
	}

	var attributes map[string]interface{}
	out, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(out, &attributes); err != nil {
		return nil, err
	}

	entity := &Deployment{
		version: VersionV1,
		name:    name,
		spec:    spec,
		state: &State{
			revision:  1,
			status:    ReadyStateStatus,
			weight:    0,
			schedule:  nil,
			retry:     nil,
			createdAt: now,
			updatedAt: now,
		},
		history: []*History{
			{
				event:      CreateEventType,
				revision:   1,
				status:     ReadyStateStatus,
				weight:     0,
				createdAt:  now,
				attributes: attributes,
			},
		},
	}

	validate := validator.New()
	if err := validate.Struct(entity); err != nil {
		return nil, err
	}

	return entity, nil
}

package deployment_test

import (
	"encoding/json"
	"github.com/k-kinzal/progressived/pkg/progressived/persistence/v1/deployment"
	"testing"
	"time"
)

func newDeployment() *deployment.Deployment {
	return &deployment.Deployment{
		Version: deployment.V1,
		Name:    "example-deployment",
		Spec: deployment.Spec{
			Interval: 60,
			Provider: &deployment.ProviderSpec{
				ProviderType: "inmemory",
				InMemory:     &deployment.InMemoryProviderSpec{},
			},
			Step: &deployment.StepBehaviorSpec{
				StepBehaviorAlgorithm: "increase",
				Increase: &deployment.IncreaseStepBehaviorSpec{
					Threshold: 25,
				},
			},
			Rollback: &deployment.RollbackBehaviorSpec{
				RollbackBehaviorAlgorithm: "history",
				History:                   &deployment.HistoryRollbackBehaviorSpec{},
			},
			Metrics: []*deployment.MetricsSpec{
				{
					MetricType:  "inmemory",
					Period:      60,
					Condition:   "x > 10",
					Query:       "",
					AllowNoData: nil,
					Target: &deployment.MetricsTargetSpec{
						Percentage: 99.5,
						TimeWindow: 600,
					},
				},
			},
		},
		State: deployment.State{
			Revision: 1,
			Status:   "ready",
			Weight:   0,
			Schedule: &deployment.ScheduleState{
				Weight:            25,
				NextScheduledTime: time.Now().Add(60 * time.Second),
			},
			Retry:     nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		History: []*deployment.History{
			{
				Event:      "READY",
				Revision:   1,
				Status:     "ready",
				Weight:     0,
				CreatedAt:  time.Now(),
				Attributes: nil,
			},
		},
	}
}

func TestDeployment_StructTag_validateSuccess(t *testing.T) {
	jsonString := `
{
	"version": "v1",
	"name": "example-deployment",
	"spec": {
	},
	"state": {
	},
	"history": []
}`
	var entity *deployment.Deployment
	if err := json.Unmarshal([]byte(jsonString), &entity); err != nil {
		t.Fatal(err)
	}

}

func TestDeployment_StructTag_jsonUnmarshalAll(t *testing.T) {
	entity := newDeployment()
	if err := deployment.Validate(entity); err != nil {
		t.Error(err)
	}
}

func TestDeployment_Clone(t *testing.T) {
	entity1 := newDeployment()
	entity2 := entity1.Clone()

	if entity1 == entity2 {
		t.Fatal("")
	}
}

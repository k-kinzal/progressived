package storage

import (
	"github.com/k-kinzal/progressived/pkg/provider"
	"time"
)

type Status string

const (
	StatusReady Status = "ready"
	StatusPause = "pause"
	StatusInProgress = "in progress"
	StatusRollback = "rollback"
	StatusRollbackCompleted = "rollback completed"
	StatusCompleted = "completed"
)

type State struct {
	Version string `json:"versin"`
	Name string `json:"name"`

	Spec StateSpec `json:"spec,omitempty"`
	Status StateStatus `json:"status,omitempty"`
}

type StateSpec struct {
	Weight float64 `json:"weight"`
	Interval time.Duration `json:"interval"`

	Provider StateProviderSpec `json:"provider,omitempty"`
}

type StateProviderSpec struct {
	Type provider.ProviderType `json:"type"`
	ID string `json:"id"`
	Config interface{} `json:"config"`
}

type StateStatus struct {
	Status Status `json:"status,omitempty"`
	Reason string `json:"reason,omitempty"`
	CreateAt time.Time `json:"createAt,omitempty"`
	UpdateAt time.Time `json:"updateAt,omitempty"`
}

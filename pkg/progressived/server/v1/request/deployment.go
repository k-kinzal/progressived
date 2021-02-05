package request

import (
	"time"
)

type ListDeploymentRequest struct {
}

type DescribeDeploymentRequest struct {
}

type PutDeploymentRequest struct {
	Interval time.Duration `json:"interval"`
	Provider struct {
		ProviderType  string `json:"type"`
		HostedZoneID  string `json:"hostedZoneId,omitempty"`
		RecordName    string `json:"recordName,omitempty"`
		RecordType    string `json:"recordType,omitempty"`
		SetIdentifier string `json:"setIdentifier,omitempty"`
	} `json:"provider"`
	Step struct {
		Algorithm string  `json:"algorithm"`
		Threshold float64 `json:"threshold,omitempty"`
	} `json:"step,omitempty"`
	Rollback struct {
		Algorithm string  `json:"algorithm"`
		Threshold float64 `json:"threshold,omitempty"`
	} `json:"rollback,omitempty"`
	Metrics []struct {
		MetricType  string        `json:"name"`
		Period      time.Duration `json:"period"`
		Condition   string        `json:"condition"`
		Query       string        `json:"query,omitempty"`
		AllowNoData *bool         `json:"allowNoData,omitempty"`
		Target      *struct {
			Percentage float64       `json:"percentage"`
			TimeWindow time.Duration `json:"timeWindow"`
		} `json:"target,omitempty"`
	} `json:"metrics,omitempty"`
	AutoStart bool `json:"autoStart,omitempty"`
}

type ScheduleDeploymentRequest struct {
	NextScheduleTime *time.Time `json:"time,omitempty"`
}

type PauseDeploymentRequest struct {
}

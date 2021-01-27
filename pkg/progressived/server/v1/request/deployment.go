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
		ProviderType  string `json:"type" validate:"required"`
		HostedZoneID  string `json:"hostedZoneId,omitempty"`
		RecordName    string `json:"recordName,omitempty"`
		RecordType    string `json:"recordType,omitempty"`
		SetIdentifier string `json:"setIdentifier,omitempty"`
	} `json:"provider" validate:"required"`
	Step struct {
		Algorithm string  `json:"algorithm" validate:"required"`
		Threshold float64 `json:"threshold,omitempty"`
	} `json:"step,omitempty"`
	Rollback struct {
		Algorithm string  `json:"algorithm" validate:"required"`
		Threshold float64 `json:"threshold,omitempty"`
	} `json:"rollback,omitempty"`
	Metrics []struct {
		MetricType  string        `json:"name" validate:"required"`
		Period      time.Duration `json:"period" valdate:"required"`
		Condition   string        `json:"condition" valdate:"required"`
		Query       string        `json:"query,omitempty"`
		AllowNoData *bool         `json:"allowNoData,omitempty"`
		Target      struct {
			Percentage float64       `json:"percentage" valdate:"required"`
			TimeWindow time.Duration `json:"timeWindow" valdate:"required"`
		} `json:"target,omitempty"`
	} `json:"metrics,omitempty"`
}

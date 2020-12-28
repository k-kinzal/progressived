package conf

import "time"

// Config

type Config struct {
	LogLevel string `yaml:"logLevel" validate:"required,oneof=debug info warr error" default:"info"`
	Rules     []RuleConfig    `yaml:"rules" validate:"required"`
}

// Rule

type RuleConfig struct {
	Interval time.Duration `yaml:"interval" validate:"required"`
	Type string `yaml:"type" validate:"required,oneof=route53"`
	Route53Provider Route53ProviderConfig `yaml:"route53,omitempty"`
	Algorithm string  `yaml:"algorithm" validate:"required,oneof=increase decrease"`
	Threshold float64 `yaml:"threshold" validate:"required"`

	Metrics   MetricsConfig   `yaml:"metrics,omitempty" validate:"required"`
}


// Providers

type Route53ProviderConfig struct {
	HostedZoneId          string `yaml:"hostedZoneId" validate:"required"`
	RecordName            string `yaml:"recordName,omitempty"`
	SourceIdentifier      string `yaml:"sourceIdentifier" validate:"required"`
	DestinationIdentifier string `yaml:"destinationIdentifier" validate:"required"`
}

// Metrics

type MetricsConfig struct {
	Type        string        `yaml:"type" validate:"required,oneof=cloudwatch"`
	Period      time.Duration `yaml:"period" default:"5m"`
	Query       string        `yaml:"query" validate:"required"`
	AllowNoData bool          `yaml:"allowNoData" default:"false"`
	Condition   string        `yaml:"condition" validate:"required"`

	CloudWatchMetricsConfig CloudWatchMetricsConfig `yaml:"cloudwatch,omitempty"`
}
type CloudWatchMetricsConfig struct {
}



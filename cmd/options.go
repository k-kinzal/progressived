package cmd

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/k-kinzal/progressived/pkg/algorithm"
	"github.com/k-kinzal/progressived/pkg/formura"
	"github.com/k-kinzal/progressived/pkg/metrics"
	"github.com/k-kinzal/progressived/pkg/provider"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

var (
	config Config
)

type Route53ProviderConfig struct {
	HostedZoneId          string `yaml:"hostedZoneId"`
	RecordName            string `yaml:"recordName"`
	SourceIdentifier      string `yaml:"sourceIdentifier"`
	DestinationIdentifier string `yaml:"destinationIdentifier"`
}

type ProviderConfig struct {
	Type string `yaml:"type"`

	Route53Provider Route53ProviderConfig `yaml:"route53"`
}

type CloudWatchMetricsConfig struct {
}

type MetricsConfig struct {
	Type        string        `yaml:"type"`
	Period      time.Duration `yaml:"period"`
	Query       string        `yaml:"query"`
	AllowNoData bool          `yaml:"allowNoData"`
	Condition   string        `yaml:"condition"`

	CloudWatchMetricsConfig CloudWatchMetricsConfig `yaml:"cloudwatch"`
}

type AlgorithmConfig struct {
	Type  string  `yaml:"type"`
	Value float64 `yaml:"value"`
}

type Config struct {
	Provider  ProviderConfig  `yaml:"provider"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	Algorithm AlgorithmConfig `yamk:"algorithm"`
}

func setFlags(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringVar(&config.Provider.Type, "provider", provider.Route53ProviderType, "The provider of the request routing policy")
	cmd.Flags().StringVar(&config.Provider.Route53Provider.HostedZoneId, "route53-hosted-zone-id", "", "Host zone ID for AWS Route53")
	cmd.Flags().StringVar(&config.Provider.Route53Provider.RecordName, "route53-record-name", "", "Record Name for AWS Route53")
	cmd.Flags().StringVar(&config.Provider.Route53Provider.SourceIdentifier, "route53-source-identifier", "", "Identifier of the AWS Route53 migration source")
	cmd.Flags().StringVar(&config.Provider.Route53Provider.DestinationIdentifier, "route53-destination-identifier", "", "Identifier of the Route53 migration destination")
	cmd.Flags().StringVar(&config.Metrics.Type, "metrics-type", metrics.CloudWatchMetricsType, "Types of metrics to collect")
	cmd.Flags().DurationVar(&config.Metrics.Period, "metrics-period", 5*time.Minute, "Collection period for metrics")
	cmd.Flags().StringVar(&config.Metrics.Query, "query", "", "A query to collect metrics")
	cmd.Flags().BoolVar(&config.Metrics.AllowNoData, "allow-no-data", false, "If true, allow the collection of metrics to no data")
	cmd.Flags().StringVar(&config.Metrics.Condition, "condition", "", "Rollback if the collected metrics do not match the conditions")
	cmd.Flags().StringVar(&config.Algorithm.Type, "algorithm", algorithm.IncreaseAlgorithm, "Algorithm for determining the value to be updated")
	cmd.Flags().Float64Var(&config.Algorithm.Value, "value", 10, "Reference value to be applied to the algorithm")

	return cmd
}

func newProvider(config Config) (provider.Provider, error) {
	var prov provider.Provider
	switch config.Provider.Type {
	case provider.Route53ProviderType:
		if config.Provider.Route53Provider.HostedZoneId == "" {
			return nil, fmt.Errorf("if the provider is \"%s\", the --route53-hosted-zone-id is required", provider.Route53ProviderType)
		}
		if config.Provider.Route53Provider.RecordName == "" {
			return nil, fmt.Errorf("if the provider is \"%s\", the --route53-recourud-name is required", provider.Route53ProviderType)
		}
		if config.Provider.Route53Provider.SourceIdentifier == "" {
			return nil, fmt.Errorf("if the provider is \"%s\", the --route53-source-identifier is required", provider.Route53ProviderType)
		}
		if config.Provider.Route53Provider.DestinationIdentifier == "" {
			return nil, fmt.Errorf("if the provider is \"%s\", the --route53-destination-identifier is required", provider.Route53ProviderType)
		}

		config := &provider.Route53Confg{
			Sess:                  awsSession,
			HostedZoneId:          config.Provider.Route53Provider.HostedZoneId,
			RecordName:            config.Provider.Route53Provider.RecordName,
			SourceIdentifier:      config.Provider.Route53Provider.SourceIdentifier,
			DestinationIdentifier: config.Provider.Route53Provider.DestinationIdentifier,
		}
		p, err := provider.NewRoute53Provider(config)
		if err != nil {
			return nil, err
		}
		prov = p
	default:
		return nil, fmt.Errorf("--provider can be either \"%s\"", provider.Route53ProviderType)
	}

	return prov, nil
}

func newMetrics(config Config) (metrics.Metrics, error) {
	var met metrics.Metrics
	switch config.Metrics.Type {
	case metrics.CloudWatchMetricsType:
		config := &metrics.CloudWatchConfig{
			Sess:   awsSession,
			Period: config.Metrics.Period,
		}
		met = metrics.NewCloudWatchMetrics(config)
	default:
		return nil, fmt.Errorf("--metrics-type can be either \"%s\"", metrics.CloudWatchMetricsType)
	}

	return met, nil
}

func newAlgorithm(config Config) (algorithm.Algorithm, error) {
	var algo algorithm.Algorithm
	switch config.Algorithm.Type {
	case algorithm.IncreaseAlgorithm:
		algo = algorithm.NewIncretion(config.Algorithm.Value)
	case algorithm.DecreaseAlgorithm:
		algo = algorithm.NewDecrease(config.Algorithm.Value)
	default:
		return nil, fmt.Errorf("--algorithm can be either \"%s\", \"%s\"", algorithm.IncreaseAlgorithm, algorithm.DecreaseAlgorithm)
	}

	return algo, nil
}

func newQueryBuilder(config Config) (*metrics.QueryBuilder, error) {
	data := structs.Map(config)

	env := make(map[string]string)
	for _, v := range os.Environ() {
		s := strings.Split(v, "=")
		if len(s) != 2 {
			continue
		}
		env[s[0]] = s[1]
	}
	data["Environment"] = env

	return metrics.NewQueryBuikder(config.Metrics.Query, data), nil
}

func newFomura(config Config) (*formura.Formula, error) {
	return formura.NewFormula(config.Metrics.Condition), nil
}

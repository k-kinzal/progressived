package metrics

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"time"
)

const (
	CloudWatchMetricsType = "cloudwatch"
)

type CloudWatchConfig struct {
	Sess *session.Session

	Period time.Duration
}

type CloudWatchMetrics struct {
	client *cloudwatch.CloudWatch
	period time.Duration
}

func (m *CloudWatchMetrics) GetMetric(query string) (float64, error) {
	var queries []*cloudwatch.MetricDataQuery
	if err := json.Unmarshal([]byte(query), &queries); err != nil {
		return 0, fmt.Errorf("unmarshal to cloudwatch.MetricDataQuery failed: %w", err)
	}

	end := time.Now()
	start := end.Add(-m.period)
	res, err := m.client.GetMetricData(&cloudwatch.GetMetricDataInput{
		EndTime:           aws.Time(end),
		MaxDatapoints:     aws.Int64(20),
		StartTime:         aws.Time(start),
		MetricDataQueries: queries,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to get cloudwatch metrics: %w", err)
	}

	metrics := res.MetricDataResults
	if len(metrics) < 1 {
		return 0, &NoDataError{query: query}
	}
	values := metrics[0].Values
	if len(values) < 1 {
		return 0, &NoDataError{query: query}
	}

	return aws.Float64Value(values[0]), nil
}

func NewCloudWatchMetrics(config *CloudWatchConfig) Metrics {
	client := cloudwatch.New(config.Sess)

	return &CloudWatchMetrics{
		client: client,
		period: config.Period,
	}
}

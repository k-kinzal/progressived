package progressived

import (
	"fmt"
	"github.com/k-kinzal/progressived/pkg/algorithm"
	"github.com/k-kinzal/progressived/pkg/formura"
	"github.com/k-kinzal/progressived/pkg/metrics"
	"github.com/k-kinzal/progressived/pkg/provider"
)

type Progressived struct {
	Provider  provider.Provider
	Metrics   metrics.Metrics
	Builder   *metrics.QueryBuilder
	Algorithm algorithm.Algorithm
	Formura   *formura.Formula

	AllowNoData bool
}

type NotMatchMetricsError struct {
	metricsValue float64
	condition    string
}

func (e NotMatchMetricsError) Error() string {
	return fmt.Sprintf("metrics value `%f` did not match the specified conditions `%s`", e.metricsValue, e.condition)
}

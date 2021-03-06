package progressived

import (
	"fmt"
	"github.com/k-kinzal/progressived/pkg/metrics"
)

func (p *Progressived) Update() (float64, error) {
	query, err := p.Builder.Build(nil)
	if err != nil {
		return -1, fmt.Errorf("update: %w", err)
	}

	value, err := p.Metrics.GetMetric(query)
	if err != nil {
		if _, ok := err.(*metrics.NoDataError); !ok || !p.AllowNoData {
			return -1, fmt.Errorf("update: %w", err)
		}
	}
	if _, ok := err.(*metrics.NoDataError); !ok {
		ok, err := p.Formura.Eval(value)
		if err != nil {
			return -1, fmt.Errorf("update: %w", err)
		}
		if !ok {
			return -1, &NotMatchMetricsError{value, p.Formura.Expression()}
		}
	}

	percentage, err := p.CurrentPercentage()
	if err != nil {
		return -1, fmt.Errorf("update: %w", err)
	}
	updatePercentage, err := p.NextPercentage()
	if err != nil {
		return -1, fmt.Errorf("update: %w", err)
	}
	if percentage == updatePercentage {
		return -1, AlreadyCompletedError{}
	}

	if err := p.Provider.Update(updatePercentage); err != nil {
		return -1, fmt.Errorf("update: %w", err)
	}

	return updatePercentage, nil
}

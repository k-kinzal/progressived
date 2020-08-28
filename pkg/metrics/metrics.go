package metrics

import "fmt"

type Metrics interface {
	GetMetric(query string) (float64, error)
}

type NoDataError struct {
	query string
}

func (e *NoDataError) Error() string {
	return fmt.Sprintf("There was no data in the `%s`", e.query)
}

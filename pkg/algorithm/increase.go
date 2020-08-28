package algorithm

const (
	IncreaseAlgorithm = "increase"
)

type Increase struct {
	value float64
}

func (a Increase) Next(value float64) float64 {
	return value + a.value
}

func (a Increase) Previous(value float64) float64 {
	return value - a.value
}

func NewIncretion(value float64) Algorithm {
	return &Increase{
		value: value,
	}
}

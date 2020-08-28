package algorithm

const (
	DecreaseAlgorithm = "decrease"
)

type Decrease struct {
	value float64
}

func (a Decrease) Next(value float64) float64 {
	return value - a.value
}

func (a Decrease) Previous(value float64) float64 {
	return value + a.value
}

func NewDecrease(value float64) Algorithm {
	return &Decrease{
		value: value,
	}
}

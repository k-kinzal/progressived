package algorithm

type Algorithm interface {
	Next(float64) float64
	Previous(float64) float64
}

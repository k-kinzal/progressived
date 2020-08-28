package provider

type Provider interface {
	Get() (float64, error)
	Update(percent float64) error
}

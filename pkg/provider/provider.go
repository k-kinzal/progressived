package provider

type ProviderType string

type Provider interface {
	TargetName() string
	Get() (float64, error)
	Update(percent float64) error
}

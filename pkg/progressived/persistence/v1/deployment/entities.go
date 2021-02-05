package deployment

import "fmt"

type Deployments interface {
	Put(entity *Deployment) error
	Get(name string) (*Deployment, error)
	Seq() ([]*Deployment, error)
	Changes() <-chan *Deployment
}

type NotFoundError struct {
	name string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s: not found", e.name)
}

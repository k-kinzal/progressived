package deployment

import (
	"github.com/k-kinzal/progressived/pkg/progressived/persistence"
	"sync"
)

//import (
//	"github.com/k-kinzal/progressived/pkg/progressived/entity"
//	"sync"
//)
//
//type DeploymentRepositoryOnInMemory struct {
//	data map[string]*entity.Deployment
//	mu   sync.Mutex
//}
//
//func (r *DeploymentRepositoryOnInMemory) AsEntities(_ *DeploymentFilter) ([]*entity.Deployment, error) {
//	e := make([]*entity.Deployment, 0)
//	for _, v := range r.data {
//		e = append(e, v)
//	}
//	return e, nil
//}
//
//func (r *DeploymentRepositoryOnInMemory) ResolveById(id string) (*entity.Deployment, error) {
//	e, ok := r.data[id]
//	if ok == false {
//		return nil, &DeploymentNotFoundError{id}
//	}
//	return e, nil
//}
//
//func (r *DeploymentRepositoryOnInMemory) Store(state *entity.Deployment) error {
//	r.mu.Lock()
//	defer r.mu.Unlock()
//
//	r.data[state.ID] = state
//
//	return nil
//}
//
//func NewDeploymentRepositoryOnInMemory() *DeploymentRepositoryOnInMemory {
//	return &DeploymentRepositoryOnInMemory{
//		data: make(map[string]*entity.Deployment, 0),
//		mu:   sync.Mutex{},
//	}
//}

type DeploymentsOnInMemory struct {
	data map[string]*Deployment
	ch   chan persistence.Deployment
	mu   sync.Mutex
}

func (d *DeploymentsOnInMemory) Put(entity persistence.Deployment) error {
	d.mu.Lock()

	e := entity.(*Deployment)

	d.data[e.Name] = e.Clone()

	d.mu.Unlock()

	d.ch <- e.Clone()

	return nil
}

func (d *DeploymentsOnInMemory) Get(name string) (persistence.Deployment, error) {
	entity, ok := d.data[name]
	if !ok {
		return nil, &persistence.DeploymentNotFountdError{Name: name}
	}
	return entity.Clone(), nil
}

func (d *DeploymentsOnInMemory) Seq() ([]persistence.Deployment, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entities := make([]persistence.Deployment, len(d.data))
	i := 0
	for _, entity := range d.data {
		entities[i] = entity.Clone()
		i++
	}
	return entities, nil
}

func (d *DeploymentsOnInMemory) Filter(filter persistence.DeploymentFilterFunc) (persistence.DeploymentsReadOnly, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	data := make(map[string]*Deployment, 0)
	for name, entity := range d.data {
		if filter(entity) {
			data[name] = entity.Clone()
		}
	}

	return &DeploymentsOnInMemory{
		data: data,
		mu:   sync.Mutex{},
	}, nil
}

func (d *DeploymentsOnInMemory) Changes() <-chan persistence.Deployment {
	return d.ch
}

func NewDeploymentsOnInMemory() *DeploymentsOnInMemory {
	return &DeploymentsOnInMemory{
		data: make(map[string]*Deployment, 0),
		mu:   sync.Mutex{},
		ch:   make(chan persistence.Deployment, 1),
	}
}

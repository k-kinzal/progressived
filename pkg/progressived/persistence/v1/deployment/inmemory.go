package deployment

import (
	"sync"
)

type DeploymentsOnInMemory struct {
	data map[string]*Deployment
	ch   chan *Deployment
	mu   sync.Mutex
}

func (d *DeploymentsOnInMemory) Put(entity *Deployment) error {
	d.mu.Lock()

	d.data[entity.name] = entity.Clone()

	d.mu.Unlock()

	d.ch <- entity.Clone()

	return nil
}

func (d *DeploymentsOnInMemory) Get(name string) (*Deployment, error) {
	entity, ok := d.data[name]
	if !ok {
		return nil, &NotFoundError{name: name}
	}
	return entity.Clone(), nil
}

func (d *DeploymentsOnInMemory) Seq() ([]*Deployment, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	entities := make([]*Deployment, len(d.data))
	i := 0
	for _, entity := range d.data {
		entities[i] = entity.Clone()
		i++
	}
	return entities, nil
}

func (d *DeploymentsOnInMemory) Changes() <-chan *Deployment {
	return d.ch
}

func NewDeploymentsOnInMemory() *DeploymentsOnInMemory {
	return &DeploymentsOnInMemory{
		data: make(map[string]*Deployment, 0),
		mu:   sync.Mutex{},
		ch:   make(chan *Deployment, 1),
	}
}

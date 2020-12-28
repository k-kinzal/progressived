package inmemory

import (
	"fmt"
	"github.com/k-kinzal/progressived/pkg/storage"
	"sync"
)

type Storage struct {
	data map[string]*storage.State
	mu   sync.Mutex
}

func (s *Storage) Write(state *storage.State) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[state.Name] = state

	return nil
}

func (s *Storage) Read(name string) (*storage.State, error) {
	if val, ok := s.data[name]; !ok {
		return nil, fmt.Errorf("storage: not found state `%s` in memory", name)
	} else {
		return val, nil
	}
}

func NewInMemoryStorage() *Storage {
	return &Storage{
		data: make(map[string]*storage.State, 0),
		mu:   sync.Mutex{},
	}
}
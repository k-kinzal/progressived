package storage

type Storage interface {
	Write(*State) error
	Read() (*State, error)
	Change() <-chan *State
}
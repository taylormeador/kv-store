package store

import "errors"

type Store struct {
	data map[string]string
}

var ErrKeyNotExists error = errors.New("key does not exist")

func (s *Store) Get(key string) (string, error) {
	val, ok := s.data[key]
	if ok {
		return val, nil
	} else {
		return val, ErrKeyNotExists
	}
}

func (s *Store) Set(key string, value string) {
	s.data[key] = value
}

func (s *Store) Delete(key string) {
	delete(s.data, key)
}

func (s *Store) Exists(key string) bool {
	_, exists := s.data[key]
	return exists
}

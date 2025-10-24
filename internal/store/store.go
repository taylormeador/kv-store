package store

type Store struct {
	data map[string]string
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]string),
	}
}

func (s *Store) Get(key string) (string, bool) {
	val, exists := s.data[key]
	return val, exists
}

func (s *Store) Set(key string, value string) {
	s.data[key] = value
}

func (s *Store) Delete(key string) bool {
	_, exists := s.data[key]
	delete(s.data, key)
	return exists
}

func (s *Store) Exists(key string) bool {
	_, exists := s.data[key]
	return exists
}

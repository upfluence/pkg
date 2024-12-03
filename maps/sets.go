package maps

import (
	"sync"
)

// Set is a concurrency safe wrapper around a map to ensure uniqueness through
// a collection.
// While the underlying map can be manually manipulated, read and
// mutation should be done with through methods to ensure
// concurrency safety.
type Set[T comparable] struct {
	mu sync.RWMutex

	Set map[T]struct{}
}

func (s *Set[T]) Add(vs ...T) {
	if len(vs) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Set == nil {
		s.Set = make(map[T]struct{}, len(vs))
	}

	for _, v := range vs {
		s.Set[v] = struct{}{}
	}
}

func (s *Set[T]) Has(v T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Set) == 0 {
		return false
	}

	_, ok := s.Set[v]

	return ok
}

func (s *Set[T]) Keys() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.Set) == 0 {
		return nil
	}

	return Keys(s.Set)
}

func (s *Set[T]) Delete(vs ...T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range vs {
		delete(s.Set, v)
	}
}

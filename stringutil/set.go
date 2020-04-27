package stringutil

import "sync"

type StringSet = Set

type Set struct {
	sync.Once

	Set map[string]struct{}
}

func (s *Set) Add(vs ...string) {
	if len(vs) == 0 {
		return
	}

	s.Do(func() {
		if s.Set == nil {
			s.Set = make(map[string]struct{}, len(vs))
		}
	})

	for _, v := range vs {
		s.Set[v] = struct{}{}
	}
}

func (s *Set) Has(v string) bool {
	if len(s.Set) == 0 {
		return false
	}

	_, ok := s.Set[v]
	return ok
}

func (s *Set) Strings() []string {
	if len(s.Set) == 0 {
		return nil
	}

	res := make([]string, 0, len(s.Set))

	for v := range s.Set {
		res = append(res, v)
	}

	return res
}

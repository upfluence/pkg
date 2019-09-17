package stringutil

import "sync"

type StringSet struct {
	Set map[string]struct{}

	sync.Once
}

func NewStringSet() *StringSet {
	return &StringSet{}
}

func (ss *StringSet) Add(nss ...string) {
	if len(nss) == 0 {
		return
	}

	ss.Do(func() { ss.Set = make(map[string]struct{}) })

	for _, ns := range nss {
		ss.Set[ns] = struct{}{}
	}
}

func (ss *StringSet) Strings() []string {
	var s []string

	for v := range ss.Set {
		s = append(s, v)
	}

	return s
}

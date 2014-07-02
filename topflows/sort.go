package main

import (
	"sort"
)

type sortable struct {
	keys []string
	m    map[string]uint64
}

func (s sortable) Len() int {
	return len(s.keys)
}

func (s sortable) Swap(i, j int) {
	s.keys[i], s.keys[j] = s.keys[j], s.keys[i]
}

func (s sortable) Less(i, j int) bool {
	return s.m[s.keys[i]] < s.m[s.keys[j]]
}

func sortMap(m map[string]uint64) []string {
	var keys []string
	for key, _ := range m {
		keys = append(keys, key)
	}

	s := sortable{
		keys: keys,
		m:    m,
	}

	sort.Sort(sort.Reverse(s))

	return s.keys
}

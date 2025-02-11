package inspect

import (
	"fmt"
	"sync"
)

var configMap = storeMap{m: map[string]snippetDetails{}}

func configKey(file string, line int) string { return fmt.Sprintf("%s:%d", file, line) }

type storeMap struct {
	m map[string]snippetDetails
	sync.RWMutex
}

func (s *storeMap) set(key string, val snippetDetails) {
	s.Lock()
	defer s.Unlock()

	s.m[key] = val
}

func (s *storeMap) get(key string) snippetDetails {
	s.RLock()
	defer s.RUnlock()

	return s.m[key]
}

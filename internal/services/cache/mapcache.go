package cache

import "sync"

type mapCache struct {
	locker sync.Mutex
	keys   map[string]struct{}
}

func NewMapCache() Cache {
	return &mapCache{
		keys: make(map[string]struct{}),
	}
}

func (m *mapCache) Lock(keys []string) error {
	m.locker.Lock()
	defer m.locker.Unlock()

	for _, k := range keys {
		_, ok := m.keys[k]
		if ok {
			return ErrExists
		}
	}

	for _, k := range keys {
		m.keys[k] = struct{}{}
	}

	return nil
}

func (m *mapCache) Unlock(keys []string) {
	m.locker.Lock()
	defer m.locker.Unlock()

	for _, k := range keys {
		delete(m.keys, k)
	}
}

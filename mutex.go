package cushion

import "sync"

type MutexWithKey[K comparable] struct {
	mu sync.Mutex
	m  map[K]*sync.Mutex
}

func NewMutexWithKey[K comparable]() *MutexWithKey[K] {
	return &MutexWithKey[K]{m: make(map[K]*sync.Mutex)}
}

func (m *MutexWithKey[K]) Get(k K) *sync.Mutex {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.m[k]; !ok {
		m.m[k] = &sync.Mutex{}
	}
	return m.m[k]
}

func (m *MutexWithKey[K]) Lock(k K) {
	m.Get(k).Lock()
}

func (m *MutexWithKey[K]) Unlock(k K) {
	m.Get(k).Unlock()
}

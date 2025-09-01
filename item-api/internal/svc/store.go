package svc

import (
	"sync"
	"time"
)

type memoryItem struct {
	value  string
	expire time.Time
}

type MemoryStore struct {
	mu   sync.Mutex
	data map[string]memoryItem
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]memoryItem)}
}

func (s *MemoryStore) Set(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = memoryItem{value: value, expire: time.Now().Add(ttl)}
}

func (s *MemoryStore) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.data[key]
	if !ok {
		return "", false
	}
	if time.Now().After(item.expire) {
		delete(s.data, key)
		return "", false
	}
	return item.value, true
}

func (s *MemoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

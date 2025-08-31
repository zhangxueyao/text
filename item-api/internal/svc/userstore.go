package svc

import (
	"errors"
	"sync"
)

type User struct {
	Mobile   string
	Password string
	Email    string
	Age      int
	Gender   string
}

type UserStore struct {
	mu    sync.RWMutex
	users map[string]User
}

func NewUserStore() *UserStore {
	return &UserStore{users: make(map[string]User)}
}

func (s *UserStore) Add(u User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.users[u.Mobile]; ok {
		return errors.New("user already exists")
	}
	s.users[u.Mobile] = u
	return nil
}

func (s *UserStore) Get(mobile string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[mobile]
	return u, ok
}

func (s *UserStore) Exists(mobile string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.users[mobile]
	return ok
}

func (s *UserStore) ValidatePassword(mobile, password string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[mobile]
	if !ok {
		return false
	}
	return u.Password == password
}

package internal

import (
	"context"
	"sync"
	"time"
)

type Store struct {
	m       map[int]User
	mu      sync.Mutex
	counter int
}

func (s *Store) Create(ctx context.Context, u User) User {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counter++
	u.ID = s.counter
	s.m[u.ID] = u
	u.CreatedAt = time.Now()
	return u
}

func (s *Store) Exists(ctx context.Context, u User) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.m[u.ID]
	return ok
}

func (s *Store) Count(ctx context.Context) int {
	return len(s.m)
}

func (s *Store) Update(ctx context.Context, u User) User {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[u.ID] = u
	return u
}

func (s *Store) FindByID(ctx context.Context, ID int) (User, bool) {
	s.mu.Lock()
	v, ok := s.m[ID]
	s.mu.Unlock()
	return v, ok
}

func (s *Store) Remove(ctx context.Context, ID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.m[ID]
	if ok {
		delete(s.m, ID)
	}
	return ok
}

func (s *Store) All(ctx context.Context, from, size int) []User {
	s.mu.Lock()
	defer s.mu.Unlock()
	res := make([]User, 0, len(s.m))
	for k, v := range s.m {
		res[k-1] = v
	}
	return res
}

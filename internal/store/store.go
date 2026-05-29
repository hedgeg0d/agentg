package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type state struct {
	Owner    int64    `json:"owner"`
	Services []string `json:"services"`
}

type Store struct {
	mu   sync.RWMutex
	path string
	st   state
}

func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	s := &Store{path: filepath.Join(dir, "state.json")}
	raw, err := os.ReadFile(s.path)
	if err == nil {
		_ = json.Unmarshal(raw, &s.st)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return s, nil
}

func (s *Store) flush() error {
	raw, err := json.MarshalIndent(s.st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0o600)
}

func (s *Store) Owner() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.st.Owner
}

func (s *Store) ClaimOwner(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.st.Owner != 0 {
		return s.st.Owner == id
	}
	s.st.Owner = id
	_ = s.flush()
	return true
}

func (s *Store) Services() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, len(s.st.Services))
	copy(out, s.st.Services)
	return out
}

func (s *Store) AddService(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.st.Services {
		if v == name {
			return false
		}
	}
	s.st.Services = append(s.st.Services, name)
	_ = s.flush()
	return true
}

func (s *Store) RemoveService(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, v := range s.st.Services {
		if v == name {
			s.st.Services = append(s.st.Services[:i], s.st.Services[i+1:]...)
			_ = s.flush()
			return true
		}
	}
	return false
}

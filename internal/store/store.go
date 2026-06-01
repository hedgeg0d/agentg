package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

type state struct {
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
	if slices.Contains(s.st.Services, name) {
		return false
	}
	s.st.Services = append(s.st.Services, name)
	_ = s.flush()
	return true
}

func (s *Store) RemoveService(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := len(s.st.Services)
	s.st.Services = slices.DeleteFunc(s.st.Services, func(v string) bool { return v == name })
	if len(s.st.Services) == n {
		return false
	}
	_ = s.flush()
	return true
}

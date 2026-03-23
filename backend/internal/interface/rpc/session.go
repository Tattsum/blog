package rpc

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SessionStore interface {
	Create(userID string, expiresAt time.Time) (token string)
	Get(token string) (userID string, ok bool)
	Delete(token string)
}

type MemSessionStore struct {
	mu      sync.RWMutex
	entries map[string]sessionEntry
}

type sessionEntry struct {
	userID    string
	expiresAt time.Time
}

func NewMemSessionStore() *MemSessionStore {
	return &MemSessionStore{entries: make(map[string]sessionEntry)}
}

func (s *MemSessionStore) Create(userID string, expiresAt time.Time) (token string) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		token = uuid.New().String()
	} else {
		token = hex.EncodeToString(b)
	}
	s.mu.Lock()
	s.entries[token] = sessionEntry{userID: userID, expiresAt: expiresAt}
	s.mu.Unlock()
	return token
}

func (s *MemSessionStore) Get(token string) (userID string, ok bool) {
	s.mu.RLock()
	ent, exists := s.entries[token]
	s.mu.RUnlock()
	if !exists || time.Now().After(ent.expiresAt) {
		return "", false
	}
	return ent.userID, true
}

func (s *MemSessionStore) Delete(token string) {
	s.mu.Lock()
	delete(s.entries, token)
	s.mu.Unlock()
}

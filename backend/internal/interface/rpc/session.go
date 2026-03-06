package rpc

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SessionStore はセッショントークンの作成・取得・削除を抽象化する。
type SessionStore interface {
	Create(userID string, expiresAt time.Time) (token string)
	Get(token string) (userID string, ok bool)
	Delete(token string)
}

// MemSessionStore はメモリ上でセッションを保持する実装。再起動で失効する。
type MemSessionStore struct {
	mu      sync.RWMutex
	entries map[string]sessionEntry
}

type sessionEntry struct {
	userID    string
	expiresAt time.Time
}

// NewMemSessionStore は MemSessionStore を返す。
func NewMemSessionStore() *MemSessionStore {
	return &MemSessionStore{entries: make(map[string]sessionEntry)}
}

// Create はランダムなトークンを生成し、userID と有効期限を紐付けて保存する。
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

// Get はトークンに対応する userID を返す。無効または期限切れなら ok=false。
func (s *MemSessionStore) Get(token string) (userID string, ok bool) {
	s.mu.RLock()
	ent, exists := s.entries[token]
	s.mu.RUnlock()
	if !exists || time.Now().After(ent.expiresAt) {
		return "", false
	}
	return ent.userID, true
}

// Delete はトークンを削除する。
func (s *MemSessionStore) Delete(token string) {
	s.mu.Lock()
	delete(s.entries, token)
	s.mu.Unlock()
}

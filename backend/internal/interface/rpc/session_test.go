package rpc

import (
	"testing"
	"time"
)

func TestMemSessionStore_CreateGetDelete(t *testing.T) {
	store := NewMemSessionStore()
	expiresAt := time.Now().Add(time.Hour)
	token := store.Create("user-1", expiresAt)
	if token == "" {
		t.Fatal("Create should return non-empty token")
	}
	userID, ok := store.Get(token)
	if !ok || userID != "user-1" {
		t.Errorf("Get: got userID=%q ok=%v, want userID=user-1 ok=true", userID, ok)
	}
	store.Delete(token)
	_, ok = store.Get(token)
	if ok {
		t.Error("Get after Delete should return ok=false")
	}
}

func TestMemSessionStore_GetExpired(t *testing.T) {
	store := NewMemSessionStore()
	expiresAt := time.Now().Add(-time.Second)
	token := store.Create("user-1", expiresAt)
	_, ok := store.Get(token)
	if ok {
		t.Error("Get for expired session should return ok=false")
	}
}

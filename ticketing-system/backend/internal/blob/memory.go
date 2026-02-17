package blob

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
)

// MemoryStore is an in-memory ObjectStore for testing.
type MemoryStore struct {
	mu    sync.RWMutex
	blobs map[string][]byte
}

// NewMemory creates an in-memory ObjectStore.
func NewMemory() *MemoryStore {
	return &MemoryStore{blobs: make(map[string][]byte)}
}

func (m *MemoryStore) Put(_ context.Context, key string, reader io.Reader, _ int64, _ string) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read data for %q: %w", key, err)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blobs[key] = data
	return nil
}

func (m *MemoryStore) Get(_ context.Context, key string) (io.ReadCloser, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.blobs[key]
	if !ok {
		return nil, fmt.Errorf("object %q not found", key)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *MemoryStore) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.blobs, key)
	return nil
}

// Len returns the number of stored objects.
func (m *MemoryStore) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.blobs)
}

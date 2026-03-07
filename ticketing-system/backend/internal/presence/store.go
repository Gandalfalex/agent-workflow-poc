package presence

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// Entry represents one user's presence within a project.
type Entry struct {
	UserID    uuid.UUID
	UserName  string
	ProjectID uuid.UUID
	View      string // "board" or "ticket"
	TicketID  *uuid.UUID
	LastSeen  time.Time
}

// StoreIface is the interface for presence stores (for testability).
type StoreIface interface {
	Upsert(e Entry)
	Delete(projectID, userID uuid.UUID)
	ListByProject(projectID uuid.UUID) []Entry
}

// Store is an in-memory presence store with TTL eviction.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry // key: projectID+":"+userID
	ttl     time.Duration
}

// New creates a new presence Store with the given TTL.
func New(ttl time.Duration) *Store {
	return &Store{
		entries: make(map[string]Entry),
		ttl:     ttl,
	}
}

func key(projectID, userID uuid.UUID) string {
	return projectID.String() + ":" + userID.String()
}

// Upsert inserts or updates a presence entry, resetting LastSeen to now.
func (s *Store) Upsert(e Entry) {
	e.LastSeen = time.Now()
	k := key(e.ProjectID, e.UserID)
	s.mu.Lock()
	s.entries[k] = e
	s.mu.Unlock()
}

// Delete removes a presence entry.
func (s *Store) Delete(projectID, userID uuid.UUID) {
	k := key(projectID, userID)
	s.mu.Lock()
	delete(s.entries, k)
	s.mu.Unlock()
}

// ListByProject returns all non-expired entries for the given project.
func (s *Store) ListByProject(projectID uuid.UUID) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cutoff := time.Now().Add(-s.ttl)
	var out []Entry
	for _, e := range s.entries {
		if e.ProjectID == projectID && e.LastSeen.After(cutoff) {
			out = append(out, e)
		}
	}
	return out
}

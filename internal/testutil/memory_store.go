package testutil

import (
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// MemoryStore is a generic in-memory store for test mocks.
// It provides common CRUD operations scoped by owner (user_id).
type MemoryStore[T any] struct {
	mu         sync.RWMutex
	items      map[uuid.UUID]T
	getOwnerID func(T) uuid.UUID
}

// NewMemoryStore creates a new in-memory store.
// getOwnerID extracts the owner UUID from an item (e.g. the UserID field).
func NewMemoryStore[T any](getOwnerID func(T) uuid.UUID) *MemoryStore[T] {
	return &MemoryStore[T]{
		items:      make(map[uuid.UUID]T),
		getOwnerID: getOwnerID,
	}
}

// Set stores an item by its ID.
func (s *MemoryStore[T]) Set(id uuid.UUID, item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[id] = item
}

// GetByIDAndOwner returns the item matching both id and ownerID.
// Returns pgx.ErrNoRows if not found or owner doesn't match.
func (s *MemoryStore[T]) GetByIDAndOwner(id, ownerID uuid.UUID) (T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	if !ok || s.getOwnerID(item) != ownerID {
		var zero T
		return zero, pgx.ErrNoRows
	}
	return item, nil
}

// CountByOwner returns the number of items belonging to the given owner.
func (s *MemoryStore[T]) CountByOwner(ownerID uuid.UUID) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int64
	for _, item := range s.items {
		if s.getOwnerID(item) == ownerID {
			count++
		}
	}
	return count
}

// ListByOwner returns items belonging to the given owner with offset/limit pagination.
func (s *MemoryStore[T]) ListByOwner(ownerID uuid.UUID, offset, limit int) []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []T
	for _, item := range s.items {
		if s.getOwnerID(item) == ownerID {
			filtered = append(filtered, item)
		}
	}

	start := offset
	if start > len(filtered) {
		return []T{}
	}
	end := min(start+limit, len(filtered))
	return filtered[start:end]
}

// Delete removes an item by its ID.
func (s *MemoryStore[T]) Delete(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, id)
}

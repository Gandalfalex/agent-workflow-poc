package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// TicketLock represents a DB-enforced edit lock on a ticket.
type TicketLock struct {
	TicketID     uuid.UUID
	LockedBy     uuid.UUID
	LockedByName string
	AcquiredAt   time.Time
	ExpiresAt    time.Time
}

func scanTicketLock(row pgx.Row) (TicketLock, error) {
	var l TicketLock
	err := row.Scan(&l.TicketID, &l.LockedBy, &l.LockedByName, &l.AcquiredAt, &l.ExpiresAt)
	return l, err
}

// AcquireTicketLock attempts to acquire a lock on the given ticket for the given user.
// Returns (lock, true, nil) if the lock was acquired by this user (new or re-acquired).
// Returns (existingLock, false, nil) if the lock is held by someone else.
func (s *Store) AcquireTicketLock(ctx context.Context, ticketID, userID uuid.UUID, userName string, ttl time.Duration) (TicketLock, bool, error) {
	intervalStr := fmt.Sprintf("%d seconds", int(ttl.Seconds()))
	lock, err := queryOne(ctx, s.db, mustSQL("ticket_locks_acquire", nil), scanTicketLock,
		ticketID, userID, userName, intervalStr)
	if err == pgx.ErrNoRows {
		// Conflict: lock held by someone else — fetch current lock
		existing, _, getErr := s.GetTicketLock(ctx, ticketID)
		if getErr != nil {
			return TicketLock{}, false, getErr
		}
		return existing, false, nil
	}
	if err != nil {
		return TicketLock{}, false, err
	}
	return lock, true, nil
}

// GetTicketLock returns the active lock for a ticket, if any.
func (s *Store) GetTicketLock(ctx context.Context, ticketID uuid.UUID) (TicketLock, bool, error) {
	lock, err := queryOne(ctx, s.db, mustSQL("ticket_locks_get", nil), scanTicketLock, ticketID)
	if err == pgx.ErrNoRows {
		return TicketLock{}, false, nil
	}
	if err != nil {
		return TicketLock{}, false, err
	}
	return lock, true, nil
}

// RenewTicketLock extends the TTL of an existing lock held by the given user.
func (s *Store) RenewTicketLock(ctx context.Context, ticketID, userID uuid.UUID, ttl time.Duration) (TicketLock, error) {
	intervalStr := fmt.Sprintf("%d seconds", int(ttl.Seconds()))
	lock, err := queryOne(ctx, s.db, mustSQL("ticket_locks_renew", nil), scanTicketLock,
		ticketID, userID, intervalStr)
	if err == pgx.ErrNoRows {
		return TicketLock{}, pgx.ErrNoRows
	}
	return lock, err
}

// ReleaseTicketLock removes the lock held by the given user on the ticket.
func (s *Store) ReleaseTicketLock(ctx context.Context, ticketID, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, mustSQL("ticket_locks_release", nil), ticketID, userID)
	return err
}

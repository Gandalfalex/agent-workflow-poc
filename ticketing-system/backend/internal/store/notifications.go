package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Notification struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	UserID      uuid.UUID
	TicketID    uuid.UUID
	TicketKey   string
	TicketTitle string
	Type        string
	Message     string
	ReadAt      *time.Time
	CreatedAt   time.Time
}

type NotificationCreateInput struct {
	ProjectID uuid.UUID
	UserID    uuid.UUID
	TicketID  uuid.UUID
	Type      string
	Message   string
}

type NotificationFilter struct {
	ProjectID  uuid.UUID
	UserID     uuid.UUID
	UnreadOnly bool
	Limit      int
}

type NotificationPreferences struct {
	MentionEnabled    bool
	AssignmentEnabled bool
}

type NotificationPreferencesUpdateInput struct {
	MentionEnabled    *bool
	AssignmentEnabled *bool
}

func (s *Store) CreateNotification(ctx context.Context, input NotificationCreateInput) (Notification, error) {
	query := mustSQL("notifications_insert", nil)
	var id uuid.UUID
	if err := s.db.QueryRow(
		ctx,
		query,
		input.ProjectID,
		input.UserID,
		input.TicketID,
		input.Type,
		input.Message,
	).Scan(&id); err != nil {
		return Notification{}, err
	}
	return s.MarkNotificationRead(ctx, id, input.ProjectID, input.UserID)
}

func (s *Store) ListNotifications(ctx context.Context, filter NotificationFilter) ([]Notification, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	query := mustSQL("notifications_list", map[string]any{
		"UnreadOnly": filter.UnreadOnly,
	})
	return queryMany(ctx, s.db, query, scanNotification, filter.ProjectID, filter.UserID, filter.Limit)
}

func (s *Store) CountUnreadNotifications(ctx context.Context, projectID, userID uuid.UUID) (int, error) {
	query := mustSQL("notifications_unread_count", nil)
	var count int
	err := s.db.QueryRow(ctx, query, projectID, userID).Scan(&count)
	return count, err
}

func (s *Store) MarkNotificationRead(ctx context.Context, id, projectID, userID uuid.UUID) (Notification, error) {
	query := mustSQL("notifications_mark_read", nil)
	return queryOne(ctx, s.db, query, scanNotification, id, projectID, userID)
}

func (s *Store) MarkAllNotificationsRead(ctx context.Context, projectID, userID uuid.UUID) (int, error) {
	query := mustSQL("notifications_mark_all_read", nil)
	cmd, err := s.db.Exec(ctx, query, projectID, userID)
	if err != nil {
		return 0, err
	}
	return int(cmd.RowsAffected()), nil
}

func (s *Store) GetNotificationPreferences(ctx context.Context, userID uuid.UUID) (NotificationPreferences, error) {
	query := mustSQL("notification_preferences_get", nil)
	var prefs NotificationPreferences
	err := s.db.QueryRow(ctx, query, userID).Scan(&prefs.MentionEnabled, &prefs.AssignmentEnabled)
	if err == pgx.ErrNoRows {
		return NotificationPreferences{
			MentionEnabled:    true,
			AssignmentEnabled: true,
		}, nil
	}
	return prefs, err
}

func (s *Store) UpdateNotificationPreferences(ctx context.Context, userID uuid.UUID, input NotificationPreferencesUpdateInput) (NotificationPreferences, error) {
	query := mustSQL("notification_preferences_upsert", nil)
	var prefs NotificationPreferences
	err := s.db.QueryRow(ctx, query, userID, input.MentionEnabled, input.AssignmentEnabled).Scan(&prefs.MentionEnabled, &prefs.AssignmentEnabled)
	return prefs, err
}

func scanNotification(row pgx.Row) (Notification, error) {
	var n Notification
	err := row.Scan(
		&n.ID,
		&n.ProjectID,
		&n.UserID,
		&n.TicketID,
		&n.TicketKey,
		&n.TicketTitle,
		&n.Type,
		&n.Message,
		&n.ReadAt,
		&n.CreatedAt,
	)
	return n, err
}

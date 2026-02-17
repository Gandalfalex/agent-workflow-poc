package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Attachment struct {
	ID             uuid.UUID
	TicketID       uuid.UUID
	Filename       string
	ContentType    string
	Size           int64
	StorageKey     string
	UploadedBy     uuid.UUID
	UploadedByName string
	CreatedAt      time.Time
}

type AttachmentCreateInput struct {
	Filename       string
	ContentType    string
	Size           int64
	StorageKey     string
	UploadedBy     uuid.UUID
	UploadedByName string
}

func (s *Store) ListAttachments(ctx context.Context, ticketID uuid.UUID) ([]Attachment, error) {
	query := mustSQL("attachments_list", nil)
	return queryMany(ctx, s.db, query, scanAttachment, ticketID)
}

func (s *Store) GetAttachment(ctx context.Context, id uuid.UUID) (Attachment, error) {
	query := mustSQL("attachments_get", nil)
	return queryOne(ctx, s.db, query, scanAttachment, id)
}

func (s *Store) CreateAttachment(ctx context.Context, ticketID uuid.UUID, input AttachmentCreateInput) (Attachment, error) {
	query := mustSQL("attachments_insert", nil)
	return queryOne(ctx, s.db, query, scanAttachment,
		ticketID,
		input.Filename,
		input.ContentType,
		input.Size,
		input.StorageKey,
		input.UploadedBy,
		input.UploadedByName,
	)
}

func (s *Store) DeleteAttachment(ctx context.Context, id uuid.UUID) error {
	query := mustSQL("attachments_delete", nil)
	return execOne(ctx, s.db, query, pgx.ErrNoRows, id)
}

func scanAttachment(row pgx.Row) (Attachment, error) {
	var a Attachment
	err := row.Scan(
		&a.ID,
		&a.TicketID,
		&a.Filename,
		&a.ContentType,
		&a.Size,
		&a.StorageKey,
		&a.UploadedBy,
		&a.UploadedByName,
		&a.CreatedAt,
	)
	return a, err
}

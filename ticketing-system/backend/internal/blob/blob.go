package blob

import (
	"context"
	"io"
)

// ObjectStore abstracts binary object storage (MinIO, S3, in-memory, etc.).
type ObjectStore interface {
	Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

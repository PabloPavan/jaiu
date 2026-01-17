package storage

import (
	"context"
	"io"
)

type ObjectStorage interface {
	Put(ctx context.Context, key string, body io.Reader, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
}

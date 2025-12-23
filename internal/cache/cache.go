package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)

	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	Delete(ctx context.Context, key string) error

	Clear(ctx context.Context) error

	Close() error
}

type ErrCacheMiss struct {
	Key string
}

func (e *ErrCacheMiss) Error() string {
	return "cache miss: key not found"
}

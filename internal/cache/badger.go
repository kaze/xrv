package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dgraph-io/badger/v4"
)

type BadgerCache struct {
	db *badger.DB
}

func NewBadgerCache(dir string) (*BadgerCache, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger database: %w", err)
	}

	return &BadgerCache{db: db}, nil
}

func (c *BadgerCache) Get(ctx context.Context, key string) ([]byte, error) {
	var value []byte

	err := c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		value, err = item.ValueCopy(nil)
		return err
	})

	if err == badger.ErrKeyNotFound {
		return nil, &ErrCacheMiss{Key: key}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get value: %w", err)
	}

	return value, nil
}

func (c *BadgerCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), value)

		if ttl > 0 {
			entry = entry.WithTTL(ttl)
		}

		return txn.SetEntry(entry)
	})

	if err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}

	return nil
}

func (c *BadgerCache) Delete(ctx context.Context, key string) error {
	err := c.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})

	if err != nil {
		return fmt.Errorf("failed to delete value: %w", err)
	}

	return nil
}

func (c *BadgerCache) Clear(ctx context.Context) error {
	err := c.db.DropAll()
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	return nil
}

func (c *BadgerCache) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

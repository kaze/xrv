package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBadgerCache_SetAndGet(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewBadgerCache(dir)
	if err != nil {
		t.Fatalf("NewBadgerCache() error = %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	key := "test-key"
	value := []byte("test-value")

	err = cache.Set(ctx, key, value, 0)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if string(got) != string(value) {
		t.Errorf("Get() = %s, want %s", string(got), string(value))
	}
}

func TestBadgerCache_GetMiss(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewBadgerCache(dir)
	if err != nil {
		t.Fatalf("NewBadgerCache() error = %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	_, err = cache.Get(ctx, "non-existent-key")
	if err == nil {
		t.Error("Expected error for cache miss, got nil")
	}

	if _, ok := err.(*ErrCacheMiss); !ok {
		t.Errorf("Expected ErrCacheMiss, got %T", err)
	}
}

func TestBadgerCache_Delete(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewBadgerCache(dir)
	if err != nil {
		t.Fatalf("NewBadgerCache() error = %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	key := "test-key"
	value := []byte("test-value")

	cache.Set(ctx, key, value, 0)
	_, err = cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	err = cache.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = cache.Get(ctx, key)
	if err == nil {
		t.Error("Expected error after delete, got nil")
	}
}

func TestBadgerCache_Clear(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewBadgerCache(dir)
	if err != nil {
		t.Fatalf("NewBadgerCache() error = %v", err)
	}
	defer cache.Close()

	ctx := context.Background()

	cache.Set(ctx, "key1", []byte("value1"), 0)
	cache.Set(ctx, "key2", []byte("value2"), 0)
	cache.Set(ctx, "key3", []byte("value3"), 0)

	err = cache.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	for _, key := range []string{"key1", "key2", "key3"} {
		_, err := cache.Get(ctx, key)
		if err == nil {
			t.Errorf("Expected error for key %s after clear, got nil", key)
		}
	}
}

func TestBadgerCache_TTL(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewBadgerCache(dir)
	if err != nil {
		t.Fatalf("NewBadgerCache() error = %v", err)
	}
	defer cache.Close()

	ctx := context.Background()
	key := "test-key"
	value := []byte("test-value")
	ttl := 1 * time.Second

	err = cache.Set(ctx, key, value, ttl)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	_, err = cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	time.Sleep(2 * time.Second)

	_, err = cache.Get(ctx, key)
	if err == nil {
		t.Error("Expected error after TTL expiration, got nil")
	}
}

func TestBadgerCache_Persistence(t *testing.T) {
	dir := t.TempDir()
	key := "test-key"
	value := []byte("test-value")

	{
		cache, err := NewBadgerCache(dir)
		if err != nil {
			t.Fatalf("NewBadgerCache() error = %v", err)
		}

		ctx := context.Background()
		err = cache.Set(ctx, key, value, 0)
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		cache.Close()
	}

	{
		cache, err := NewBadgerCache(dir)
		if err != nil {
			t.Fatalf("NewBadgerCache() error = %v", err)
		}
		defer cache.Close()

		ctx := context.Background()
		got, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get() error after reopening = %v", err)
		}

		if string(got) != string(value) {
			t.Errorf("Get() = %s, want %s", string(got), string(value))
		}
	}
}

func TestBadgerCache_CreateDirectory(t *testing.T) {
	tempDir := t.TempDir()
	dir := filepath.Join(tempDir, "subdir", "cache")

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatal("Directory should not exist yet")
	}

	cache, err := NewBadgerCache(dir)
	if err != nil {
		t.Fatalf("NewBadgerCache() error = %v", err)
	}
	defer cache.Close()

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Directory was not created")
	}
}

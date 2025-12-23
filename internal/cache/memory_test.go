package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "test-key"
	value := []byte("test-value")

	err := cache.Set(ctx, key, value, 0)
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

func TestMemoryCache_GetMiss(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	ctx := context.Background()

	_, err := cache.Get(ctx, "non-existent-key")
	if err == nil {
		t.Error("Expected error for cache miss, got nil")
	}

	if _, ok := err.(*ErrCacheMiss); !ok {
		t.Errorf("Expected ErrCacheMiss, got %T", err)
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "test-key"
	value := []byte("test-value")

	cache.Set(ctx, key, value, 0)
	_, err := cache.Get(ctx, key)
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

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	ctx := context.Background()

	cache.Set(ctx, "key1", []byte("value1"), 0)
	cache.Set(ctx, "key2", []byte("value2"), 0)
	cache.Set(ctx, "key3", []byte("value3"), 0)

	err := cache.Clear(ctx)
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

func TestMemoryCache_TTL(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "test-key"
	value := []byte("test-value")
	ttl := 100 * time.Millisecond

	err := cache.Set(ctx, key, value, ttl)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	_, err = cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	_, err = cache.Get(ctx, key)
	if err == nil {
		t.Error("Expected error after TTL expiration, got nil")
	}
}

func TestMemoryCache_UpdateValue(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "test-key"
	value1 := []byte("value1")
	value2 := []byte("value2")

	cache.Set(ctx, key, value1, 0)

	cache.Set(ctx, key, value2, 0)

	got, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if string(got) != string(value2) {
		t.Errorf("Get() = %s, want %s", string(got), string(value2))
	}
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			key := "key"
			value := []byte("value")
			cache.Set(ctx, key, value, 0)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	_, err := cache.Get(ctx, "key")
	if err != nil {
		t.Errorf("Get() error after concurrent writes = %v", err)
	}
}

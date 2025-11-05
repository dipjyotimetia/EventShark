package idempotency

import (
	"testing"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/config"
)

func TestNewIdempotencyManager(t *testing.T) {
	cfg := &config.IdempotencyConfig{
		Enabled:    true,
		CacheTTL:   1 * time.Hour,
		MaxEntries: 1000,
	}

	manager := NewIdempotencyManager(cfg)

	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}

	if manager.cache == nil {
		t.Error("Expected cache to be initialized")
	}

	if manager.ttl != cfg.CacheTTL {
		t.Error("Expected TTL to match config")
	}

	if manager.maxEntries != cfg.MaxEntries {
		t.Error("Expected maxEntries to match config")
	}
}

func TestIdempotencyManager_Check(t *testing.T) {
	cfg := &config.IdempotencyConfig{
		Enabled:    true,
		CacheTTL:   1 * time.Hour,
		MaxEntries: 1000,
	}

	manager := NewIdempotencyManager(cfg)

	// Check non-existent key
	exists, result := manager.Check("nonexistent")
	if exists {
		t.Error("Expected key to not exist")
	}
	if result != nil {
		t.Error("Expected nil result for non-existent key")
	}

	// Store a key
	manager.Store("test-key", "test-result")

	// Check existing key
	exists, result = manager.Check("test-key")
	if !exists {
		t.Error("Expected key to exist")
	}
	if result != "test-result" {
		t.Errorf("Expected result 'test-result', got %v", result)
	}
}

func TestIdempotencyManager_Store(t *testing.T) {
	cfg := &config.IdempotencyConfig{
		Enabled:    true,
		CacheTTL:   1 * time.Hour,
		MaxEntries: 1000,
	}

	manager := NewIdempotencyManager(cfg)

	err := manager.Store("key1", "value1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	exists, value := manager.Check("key1")
	if !exists {
		t.Error("Expected key to exist after store")
	}
	if value != "value1" {
		t.Errorf("Expected value 'value1', got %v", value)
	}
}

func TestIdempotencyManager_GenerateKey(t *testing.T) {
	cfg := &config.IdempotencyConfig{
		Enabled:    true,
		CacheTTL:   1 * time.Hour,
		MaxEntries: 1000,
	}

	manager := NewIdempotencyManager(cfg)

	data := []byte("test data")
	key := manager.GenerateKey(data)

	if key == "" {
		t.Error("Expected non-empty key")
	}

	// Same data should generate same key
	key2 := manager.GenerateKey(data)
	if key != key2 {
		t.Error("Expected same key for same data")
	}

	// Different data should generate different key
	key3 := manager.GenerateKey([]byte("different data"))
	if key == key3 {
		t.Error("Expected different keys for different data")
	}
}

func TestIdempotencyManager_Expiration(t *testing.T) {
	cfg := &config.IdempotencyConfig{
		Enabled:    true,
		CacheTTL:   100 * time.Millisecond, // Short TTL for testing
		MaxEntries: 1000,
	}

	manager := NewIdempotencyManager(cfg)

	// Store a key
	manager.Store("expiring-key", "value")

	// Should exist immediately
	exists, _ := manager.Check("expiring-key")
	if !exists {
		t.Error("Expected key to exist immediately after store")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should not exist after TTL
	exists, _ = manager.Check("expiring-key")
	if exists {
		t.Error("Expected key to have expired")
	}
}

func TestIdempotencyManager_MaxEntries(t *testing.T) {
	cfg := &config.IdempotencyConfig{
		Enabled:    true,
		CacheTTL:   1 * time.Hour,
		MaxEntries: 10, // Small limit for testing
	}

	manager := NewIdempotencyManager(cfg)

	// Fill cache beyond limit
	for i := 0; i < 15; i++ {
		key := string(rune('a' + i))
		manager.Store(key, i)
	}

	// Cache should not exceed max entries significantly
	stats := manager.GetStats()
	cacheSize := stats["cache_size"].(int)
	if cacheSize > cfg.MaxEntries*2 {
		t.Errorf("Cache size %d significantly exceeds max entries %d", cacheSize, cfg.MaxEntries)
	}
}

func TestIdempotencyManager_GetStats(t *testing.T) {
	cfg := &config.IdempotencyConfig{
		Enabled:    true,
		CacheTTL:   1 * time.Hour,
		MaxEntries: 1000,
	}

	manager := NewIdempotencyManager(cfg)

	// Add some entries
	manager.Store("key1", "value1")
	manager.Store("key2", "value2")

	stats := manager.GetStats()

	if stats["cache_size"] != 2 {
		t.Errorf("Expected cache_size 2, got %v", stats["cache_size"])
	}

	if stats["max_entries"] != 1000 {
		t.Errorf("Expected max_entries 1000, got %v", stats["max_entries"])
	}

	if stats["ttl_seconds"] != 3600.0 {
		t.Errorf("Expected ttl_seconds 3600, got %v", stats["ttl_seconds"])
	}
}

func TestIdempotencyManager_ValidateIdempotencyKey(t *testing.T) {
	cfg := &config.IdempotencyConfig{
		Enabled:    true,
		CacheTTL:   1 * time.Hour,
		MaxEntries: 1000,
	}

	manager := NewIdempotencyManager(cfg)

	// Empty key should not error (optional)
	err := manager.ValidateIdempotencyKey("")
	if err != nil {
		t.Errorf("Expected no error for empty key, got %v", err)
	}

	// New key should not error
	err = manager.ValidateIdempotencyKey("new-key")
	if err != nil {
		t.Errorf("Expected no error for new key, got %v", err)
	}

	// Store the key
	manager.Store("duplicate-key", "value")

	// Duplicate key should error
	err = manager.ValidateIdempotencyKey("duplicate-key")
	if err == nil {
		t.Error("Expected error for duplicate key")
	}
}

func TestIdempotencyManager_Cleanup(t *testing.T) {
	cfg := &config.IdempotencyConfig{
		Enabled:    true,
		CacheTTL:   50 * time.Millisecond,
		MaxEntries: 1000,
	}

	manager := NewIdempotencyManager(cfg)

	// Add entries
	manager.Store("key1", "value1")
	manager.Store("key2", "value2")

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Trigger cleanup manually
	manager.cleanup()

	// Check cache is empty
	stats := manager.GetStats()
	if stats["cache_size"] != 0 {
		t.Errorf("Expected cache to be empty after cleanup, got size %v", stats["cache_size"])
	}
}

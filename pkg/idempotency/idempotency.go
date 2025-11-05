// Package idempotency provides idempotency key management and deduplication
package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/dipjyotimetia/event-shark/pkg/errors"
)

// CacheEntry represents a cached idempotency key entry
type CacheEntry struct {
	Key       string
	Result    interface{}
	CreatedAt time.Time
}

// IdempotencyManager manages idempotency keys and deduplication
type IdempotencyManager struct {
	mu         sync.RWMutex
	cache      map[string]*CacheEntry
	ttl        time.Duration
	maxEntries int
}

// NewIdempotencyManager creates a new idempotency manager
func NewIdempotencyManager(cfg *config.IdempotencyConfig) *IdempotencyManager {
	manager := &IdempotencyManager{
		cache:      make(map[string]*CacheEntry),
		ttl:        cfg.CacheTTL,
		maxEntries: cfg.MaxEntries,
	}

	// Start cleanup goroutine
	go manager.cleanupLoop()

	return manager
}

// Check checks if an idempotency key has been seen before
func (im *IdempotencyManager) Check(key string) (bool, interface{}) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	entry, exists := im.cache[key]
	if !exists {
		return false, nil
	}

	// Check if entry has expired
	if time.Since(entry.CreatedAt) > im.ttl {
		return false, nil
	}

	return true, entry.Result
}

// Store stores an idempotency key with its result
func (im *IdempotencyManager) Store(key string, result interface{}) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	// Check if cache is full
	if len(im.cache) >= im.maxEntries {
		// Remove oldest entries
		im.evictOldest()
	}

	im.cache[key] = &CacheEntry{
		Key:       key,
		Result:    result,
		CreatedAt: time.Now(),
	}

	return nil
}

// GenerateKey generates an idempotency key from request data
func (im *IdempotencyManager) GenerateKey(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// evictOldest removes the oldest entries from the cache
func (im *IdempotencyManager) evictOldest() {
	// Remove entries older than TTL
	now := time.Now()
	for key, entry := range im.cache {
		if now.Sub(entry.CreatedAt) > im.ttl {
			delete(im.cache, key)
		}
	}

	// If still over limit, remove oldest entries
	if len(im.cache) >= im.maxEntries {
		// Simple eviction: remove 10% of entries
		removeCount := im.maxEntries / 10
		count := 0
		for key := range im.cache {
			delete(im.cache, key)
			count++
			if count >= removeCount {
				break
			}
		}
	}
}

// cleanupLoop periodically cleans up expired entries
func (im *IdempotencyManager) cleanupLoop() {
	ticker := time.NewTicker(im.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		im.cleanup()
	}
}

// cleanup removes expired entries
func (im *IdempotencyManager) cleanup() {
	im.mu.Lock()
	defer im.mu.Unlock()

	now := time.Now()
	for key, entry := range im.cache {
		if now.Sub(entry.CreatedAt) > im.ttl {
			delete(im.cache, key)
		}
	}
}

// GetStats returns cache statistics
func (im *IdempotencyManager) GetStats() map[string]interface{} {
	im.mu.RLock()
	defer im.mu.RUnlock()

	return map[string]interface{}{
		"cache_size":  len(im.cache),
		"max_entries": im.maxEntries,
		"ttl_seconds": im.ttl.Seconds(),
	}
}

// ValidateIdempotencyKey validates and handles idempotency
func (im *IdempotencyManager) ValidateIdempotencyKey(key string) error {
	if key == "" {
		return nil // Idempotency key is optional
	}

	// Check if key has been seen before
	exists, _ := im.Check(key)
	if exists {
		return errors.NewDuplicateError("Request with this idempotency key has already been processed")
	}

	return nil
}

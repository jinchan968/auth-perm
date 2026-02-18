package cache

import (
	"context"
	"time"
)

// Cache is the generic interface for a cache system.
// It defines the basic operations that a cache should support.
type Cache interface {
	// Get retrieves an item from the cache.
	// It returns the item's value or an error if the key is not found.
	Get(ctx context.Context, key string) (interface{}, error)

	// Set adds an item to the cache, overwriting any existing item.
	// ttl is the time-to-live for the item. If ttl is 0, the item never expires.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes an item from the cache.
	Delete(ctx context.Context, key string) error

	// Exists checks if an item exists in the cache.
	Exists(ctx context.Context, key string) (bool, error)
}

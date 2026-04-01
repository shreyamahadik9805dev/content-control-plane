package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type TTL struct {
	inner *gocache.Cache
	ttl   time.Duration
}

// New wraps go-cache with a fixed TTL for every Set (cleanupInterval drives eviction sweeps).
func New(defaultTTL, cleanupInterval time.Duration) *TTL {
	return &TTL{
		inner: gocache.New(defaultTTL, cleanupInterval),
		ttl:   defaultTTL,
	}
}

// Get returns a cached value if it's still alive.
func (t *TTL) Get(key string) (any, bool) {
	return t.inner.Get(key)
}

func (t *TTL) Set(key string, v any) {
	t.inner.Set(key, v, t.ttl)
}

// Delete drops one key (used after writes that would make cached reads stale).
func (t *TTL) Delete(key string) {
	t.inner.Delete(key)
}

// Flush clears the whole cache (handy if we add an admin reset later).
func (t *TTL) Flush() {
	t.inner.Flush()
}

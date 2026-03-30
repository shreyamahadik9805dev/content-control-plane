package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type TTL struct {
	inner *gocache.Cache
	ttl   time.Duration
}

func New(defaultTTL, cleanupInterval time.Duration) *TTL {
	return &TTL{
		inner: gocache.New(defaultTTL, cleanupInterval),
		ttl:   defaultTTL,
	}
}

func (t *TTL) Get(key string) (any, bool) {
	return t.inner.Get(key)
}

func (t *TTL) Set(key string, v any) {
	t.inner.Set(key, v, t.ttl)
}

func (t *TTL) Delete(key string) {
	t.inner.Delete(key)
}

func (t *TTL) Flush() {
	t.inner.Flush()
}

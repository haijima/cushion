package cushion

import (
	"context"
	"sync"
	"time"
)

type Cache[K comparable, V any] struct {
	mu    *MutexWithKey[K]
	stats Stats
	*cacheConfig[K, V]
}

type cacheConfig[K comparable, V any] struct {
	expire time.Duration
	fetch  FetchFunc[K, V]
	values sync.Map
}

type FetchFunc[K comparable, V any] func(context.Context, K) (V, error)

type Value[V any] struct {
	v      V
	stored time.Time
}

func New[K comparable, V any](fetchFunc FetchFunc[K, V], expiration time.Duration) Cache[K, V] {
	return Cache[K, V]{
		mu: NewMutexWithKey[K](),
		cacheConfig: &cacheConfig[K, V]{
			expire: expiration,
			fetch:  fetchFunc,
		},
	}
}

// Get returns the value from the cache if exists and not expired.
// If not exists or expired, it fetches the value by the fetch function.
func (c *Cache[K, V]) Get(ctx context.Context, k K) (V, error) {
	c.mu.Lock(k)
	defer c.mu.Unlock(k)
	if v, exists := c.values.Load(k); exists {
		if v, ok := v.(Value[V]); ok {
			if time.Since(v.stored) < c.expire {
				c.stats.Hit()
				return v.v, nil
			}
		}
	}

	// not exists or expired
	c.stats.Miss()
	stored := time.Now()
	v, err := c.fetch(context.WithoutCancel(ctx), k)
	if err != nil {
		return v, err
	}
	c.values.Store(k, Value[V]{v: v, stored: stored})
	return v, err
}

// Warmup stores the value in advance.
func (c *Cache[K, V]) Warmup(k K, v V) {
	c.mu.Lock(k)
	defer c.mu.Unlock(k)
	c.values.Store(k, Value[V]{v: v, stored: time.Now()})
}

// Clear removes all values from the cache and resets the statistics.
func (c *Cache[K, V]) Clear() {
	c.values.Clear()
	c.stats.Reset()
}

func (c *Cache[K, V]) Stats() string {
	return c.stats.String()
}

func (c *Cache[K, V]) HitRate() float64 {
	return c.stats.HitRate()
}

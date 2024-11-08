package cushion

import (
	"context"
	"time"
)

type Cache[K comparable, V any] struct {
	mu *MutexWithKey[K]
	cacheConfig[K, V]
}

type cacheConfig[K comparable, V any] struct {
	expire time.Duration
	fetch  FetchFunc[K, V]
	values map[K]Value[V]
}

type FetchFunc[K comparable, V any] func(context.Context, K) (V, error)

type Value[V any] struct {
	v      V
	stored time.Time
}

func New[K comparable, V any](fetchFunc FetchFunc[K, V], opts ...CacheOption[K, V]) Cache[K, V] {
	config := cacheConfig[K, V]{
		expire: 5 * time.Minute,
		fetch:  fetchFunc,
		values: make(map[K]Value[V]),
	}

	for _, opt := range opts {
		config = opt(&config)
	}

	return Cache[K, V]{
		mu:          NewMutexWithKey[K](),
		cacheConfig: config,
	}
}

func (c *Cache[K, V]) Get(ctx context.Context, k K) (V, error) {
	c.mu.Lock(k)
	defer c.mu.Unlock(k)
	if v, exists := c.values[k]; exists {
		if time.Since(v.stored) < c.expire {
			return v.v, nil
		}
	}

	// not exists or expired
	stored := time.Now()
	v, err := c.fetch(context.WithoutCancel(ctx), k)
	if err != nil {
		return v, err
	}
	c.values[k] = Value[V]{v: v, stored: stored}
	return v, err
}

type CacheOption[K comparable, V any] func(*cacheConfig[K, V]) cacheConfig[K, V]

func WithExpiration[K comparable, V any](expire time.Duration) CacheOption[K, V] {
	return func(c *cacheConfig[K, V]) cacheConfig[K, V] {
		c.expire = expire
		return *c
	}
}

func WithInitialValues[K comparable, V any](values map[K]Value[V]) CacheOption[K, V] {
	return func(c *cacheConfig[K, V]) cacheConfig[K, V] {
		c.values = values
		return *c
	}
}

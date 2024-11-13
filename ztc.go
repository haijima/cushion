package cushion

// Forked from https://github.com/methane/zerotimecache

import (
	"context"
	"sync"
	"time"
)

type ZTC[V any] struct {
	m   sync.Mutex
	t   time.Time
	res V
	err error
}

func (z *ZTC[V]) DoDelay(ctx context.Context, delay time.Duration, f func(context.Context) (V, error)) (V, error) {
	t0 := time.Now()
	z.m.Lock()
	defer z.m.Unlock()

	if z.t.After(t0) {
		return z.res, z.err
	}
	if delay > 0 {
		time.Sleep(delay)
	}

	z.t = time.Now()
	z.res, z.err = f(ctx)
	return z.res, z.err
}

func (z *ZTC[V]) Do(ctx context.Context, f func(context.Context) (V, error)) (V, error) {
	return z.DoDelay(ctx, 0, f)
}

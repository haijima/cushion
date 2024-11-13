package cushion_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/haijima/cushion"
	"github.com/stretchr/testify/assert"
)

type counter struct {
	count int
}

func (c *counter) Inc(_ context.Context) (int, error) {
	c.count++
	return c.count, nil
}

func (c *counter) HeavyInc(_ context.Context) (int, error) {
	time.Sleep(50 * time.Millisecond)
	c.count++
	return c.count, nil
}

func (c *counter) Count() int {
	return c.count
}

func TestZTC_DoDelay_inDelayedDuration(t *testing.T) {
	var ztc cushion.ZTC[int]
	var c counter
	ctx := context.Background()
	wg := &sync.WaitGroup{}

	start := time.Now()
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(ctx context.Context) {
			_, _ = ztc.DoDelay(ctx, 100*time.Millisecond, c.Inc)
			wg.Done()
		}(ctx)
	}
	wg.Wait()

	elapsed := time.Since(start)
	assert.Equal(t, 1, c.Count())
	assert.Less(t, elapsed.Microseconds(), int64(105*1000)) // 100ms + buffer
}

func TestZTC_DoDelay_notInDelayedDuration(t *testing.T) {
	var ztc cushion.ZTC[int]
	var c counter
	ctx := context.Background()
	wg := &sync.WaitGroup{}

	start := time.Now()
	wg.Add(2)
	go func(ctx context.Context) {
		_, _ = ztc.DoDelay(ctx, 100*time.Millisecond, c.HeavyInc)
		wg.Done()
	}(ctx)
	time.Sleep(110 * time.Millisecond)
	go func(ctx context.Context) {
		_, _ = ztc.DoDelay(ctx, 100*time.Millisecond, c.HeavyInc)
		wg.Done()
	}(ctx)
	wg.Wait()

	elapsed := time.Since(start)
	assert.Equal(t, 2, c.Count())
	assert.Less(t, elapsed.Microseconds(), int64(305*1000))    // 300ms + buffer
	assert.Greater(t, elapsed.Microseconds(), int64(295*1000)) // 300ms - buffer
}

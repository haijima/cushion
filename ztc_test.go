package cushion_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/haijima/cushion"
)

type counter struct {
	Count int
}

func (c *counter) Inc(_ context.Context) (int, error) {
	time.Sleep(50 * time.Millisecond)
	c.Count++
	return c.Count, nil
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
	if c.Count != 1 {
		t.Errorf("expected 1, but got %d", c.Count)
	}
	if elapsed.Microseconds() >= 155*1000 { // 150ms + buffer
		t.Errorf("expected less than 155ms, but got %d", elapsed.Microseconds())
	}
}

func TestZTC_DoDelay_notInDelayedDuration(t *testing.T) {
	var ztc cushion.ZTC[int]
	var c counter
	ctx := context.Background()
	wg := &sync.WaitGroup{}

	start := time.Now()
	wg.Add(2)
	go func(ctx context.Context) {
		_, _ = ztc.DoDelay(ctx, 100*time.Millisecond, c.Inc)
		wg.Done()
	}(ctx)
	time.Sleep(100 * time.Millisecond)
	go func(ctx context.Context) {
		_, _ = ztc.DoDelay(ctx, 100*time.Millisecond, c.Inc)
		wg.Done()
	}(ctx)
	wg.Wait()

	elapsed := time.Since(start)
	if c.Count != 2 {
		t.Errorf("expected 2, but got %d", c.Count)
	}
	if elapsed.Microseconds() >= 305*1000 { // 300ms + buffer
		t.Errorf("expected less than 305ms, but got %d", elapsed.Microseconds())
	}
	if elapsed.Microseconds() <= 295*1000 { // 300ms - buffer
		t.Errorf("expected more than 295ms, but got %d", elapsed.Microseconds())
	}
}

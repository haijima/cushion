package cushion_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/haijima/cushion"
)

func TestStats_Hit(t *testing.T) {
	c := cushion.New(func(_ context.Context, k int) (int, error) { return k, nil }, 100*time.Millisecond)

	ctx := context.Background()
	wg := &sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			_, _ = c.Get(ctx, 1)
			wg.Done()
		}()
	}
	wg.Wait()

	if c.HitRate() != 0.9 {
		t.Errorf("expected 0.9, but got %f", c.HitRate())
	}
}

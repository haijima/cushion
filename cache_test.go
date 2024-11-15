package cushion_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/haijima/cushion"
)

func TestCache_Get(t *testing.T) {
	mu := &sync.Mutex{}
	data := map[int]int{1: 1, 2: 2}

	notFound := errors.New("not found")
	c := cushion.New(func(_ context.Context, k int) (int, error) {
		mu.Lock()
		defer mu.Unlock()
		res, exists := data[k]
		if exists {
			data[k] = res + 1
			return res, nil
		}
		return 0, notFound
	}, cushion.WithExpiration[int, int](100*time.Millisecond))

	ctx := context.Background()
	wg := &sync.WaitGroup{}

	for i := 0; i < 3; i++ {
		wg.Add(3)
		go func() {
			get, err := c.Get(ctx, 1)
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if get != 1 {
				t.Errorf("expected 1, but got %d", get)
			}
			wg.Done()
		}()
		go func() {
			get, err := c.Get(ctx, 2)
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if get != 2 {
				t.Errorf("expected 2, but got %d", get)
			}
			wg.Done()
		}()
		go func() {
			_, err := c.Get(ctx, 3)
			if !errors.Is(err, notFound) {
				t.Errorf("expected not found error, but got %+v", err)
			}
			wg.Done()
		}()
	}

	time.Sleep(100 * time.Millisecond)
	wg.Wait()

	for i := 0; i < 3; i++ {
		wg.Add(3)
		go func() {
			get, err := c.Get(ctx, 1)
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if get != 2 {
				t.Errorf("expected 2, but got %d", get)
			}
			wg.Done()
		}()
		go func() {
			get, err := c.Get(ctx, 2)
			if err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if get != 3 {
				t.Errorf("expected 3, but got %d", get)
			}
			wg.Done()
		}()
		go func() {
			_, err := c.Get(ctx, 3)
			if !errors.Is(err, notFound) {
				t.Errorf("expected not found error, but got %+v", err)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

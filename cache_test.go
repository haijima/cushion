package cushion_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/haijima/cushion"
)

func ExampleCache() {
	cnt := atomic.Int32{}
	var heavyFunc = func(_ context.Context, k int) (int, error) {
		// heavy operation
		time.Sleep(10 * time.Millisecond)
		cnt.Add(1)
		return k, nil
	}
	c := cushion.New(heavyFunc, 50*time.Millisecond)
	ctx := context.Background()
	wg := &sync.WaitGroup{}

	// 10 goroutines try to get the same key
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(ctx context.Context) {
			_, _ = c.Get(ctx, 1)
			wg.Done()
		}(ctx)
	}
	wg.Wait()

	fmt.Println(cnt.Load()) // heavyFunc is called only once
	// Output: 1
}

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
	}, 100*time.Millisecond)

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

func TestCache_ParallelFetch(t *testing.T) {
	c := cushion.New(func(_ context.Context, k int) (int, error) {
		time.Sleep(50 * time.Millisecond)
		return k, nil
	}, 100*time.Millisecond)

	ctx := context.Background()
	wg := &sync.WaitGroup{}

	start := time.Now()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			_, _ = c.Get(ctx, 1)
			wg.Done()
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	if elapsed.Microseconds() >= 55*1000 { // 50ms + buffer
		t.Errorf("expected less than 55ms, but got %d", elapsed.Microseconds())
	}
	if elapsed.Microseconds() <= 45*1000 { // 50ms - buffer
		t.Errorf("expected more than 45ms, but got %d", elapsed.Microseconds())
	}
}

package cushion_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/haijima/cushion"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_Get(t *testing.T) {
	mu := &sync.Mutex{}
	data := make(map[int]int)
	data = map[int]int{1: 1, 2: 2}

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
			require.NoError(t, err)
			assert.Equal(t, 1, get)
			wg.Done()
		}()
		go func() {
			get, err := c.Get(ctx, 2)
			require.NoError(t, err)
			assert.Equal(t, 2, get)
			wg.Done()
		}()
		go func() {
			_, err := c.Get(ctx, 3)
			assert.Equal(t, notFound, err)
			wg.Done()
		}()
	}

	time.Sleep(100 * time.Millisecond)
	wg.Wait()

	for i := 0; i < 3; i++ {
		wg.Add(3)
		go func() {
			get, err := c.Get(ctx, 1)
			require.NoError(t, err)
			assert.Equal(t, 2, get)
			wg.Done()
		}()
		go func() {
			get, err := c.Get(ctx, 2)
			require.NoError(t, err)
			assert.Equal(t, 3, get)
			wg.Done()
		}()
		go func() {
			_, err := c.Get(ctx, 3)
			assert.Equal(t, notFound, err)
			wg.Done()
		}()
	}

	wg.Wait()
}

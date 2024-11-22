package cushion

import (
	"fmt"
	"sync/atomic"
)

type Stats struct {
	Hits   atomic.Int32
	Misses atomic.Int32
}

func (s *Stats) Hit() {
	s.Hits.Add(1)
}

func (s *Stats) Miss() {
	s.Misses.Add(1)
}

func (s *Stats) HitCount() int32 {
	return s.Hits.Load()
}

func (s *Stats) MissCount() int32 {
	return s.Misses.Load()
}

func (s *Stats) HitRate() float64 {
	hits := s.Hits.Load()
	misses := s.Misses.Load()
	if hits+misses == 0 {
		return 0
	}
	return float64(hits) / float64(hits+misses)
}

func (s *Stats) String() string {
	return fmt.Sprintf("Hits: %d, Misses: %d, HitRate: %.2f", s.Hits.Load(), s.Misses.Load(), s.HitRate())
}

func (s *Stats) Reset() {
	s.Hits.Store(0)
	s.Misses.Store(0)
}

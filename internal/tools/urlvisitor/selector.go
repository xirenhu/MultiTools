package urlvisitor

import (
	"math/rand"
	"sync"
)

type Target struct {
	URL string
}

type TargetSelector interface {
	Next() Target
}

func NewTargetSelector(strategy string, targets []Target) TargetSelector {
	switch strategy {
	case "random":
		return &randomSelector{targets: targets}
	default:
		return &roundRobinSelector{targets: targets}
	}
}

type roundRobinSelector struct {
	mu      sync.Mutex
	index   int
	targets []Target
}

func (s *roundRobinSelector) Next() Target {
	s.mu.Lock()
	defer s.mu.Unlock()

	target := s.targets[s.index%len(s.targets)]
	s.index++
	return target
}

type randomSelector struct {
	mu      sync.Mutex
	targets []Target
}

func (s *randomSelector) Next() Target {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.targets[rand.Intn(len(s.targets))]
}

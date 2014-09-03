package balancer

import (
	"math"
	"math/rand"
	"time"
)

type Pool struct {
	backends []*Backend
	mode     BalanceMode
}

// NewPool creates a new backend pool
// Accepts a list of backend options
func newPool(backends []Backend, mode BalanceMode) *Pool {
	pool := &Pool{
		backends: make([]*Backend, len(backends)),
		mode:     mode,
	}
	for i, b := range backends {
		backend := (&b).normalize()
		pool.backends[i] = backend
		backend.start()
	}
	return pool
}

// Next returns the next best backend network/address
func (p *Pool) Next() (string, string) {
	var backend *Backend

	// Select backend
	switch p.mode {
	case ModeLeastConn:
		backend = p.nextLeastConn()
	case ModeFirstUp:
		backend = p.nextFirstUp()
	case ModeLatency:
		backend = p.nextByLatency()
	}

	// Fall back on random backend
	if backend == nil {
		backend = p.backends[rand.Intn(len(p.backends))]
	}

	// Increment the number of connections
	backend.inc()
	return backend.Network, backend.Addr
}

// Close closes the pool
func (p *Pool) Close() error {
	for _, b := range p.backends {
		b.close()
	}
	return nil
}

func (p *Pool) nextFirstUp() *Backend {
	for _, backend := range p.backends {
		if backend.Up() {
			return backend
		}
	}
	return nil
}

func (p *Pool) nextByLatency() *Backend {
	maxlcy := time.Duration(math.MaxInt64)
	pos := -1
	for n, backend := range p.backends {
		if backend.Down() {
			continue
		} else if lcy := backend.Latency(); lcy < maxlcy {
			pos, maxlcy = n, lcy
		}
	}

	if pos < 0 {
		return nil
	}
	return p.backends[pos]
}

func (p *Pool) nextLeastConn() *Backend {
	numconn := uint64(math.MaxUint64)
	pos := -1
	for n, backend := range p.backends {
		if backend.Down() {
			continue
		} else if num := backend.Connections(); num < numconn {
			pos, numconn = n, num
		}
	}

	if pos < 0 {
		return nil
	}
	return p.backends[pos]
}

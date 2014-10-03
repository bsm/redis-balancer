package balancer

type Pool struct {
	backends Backends
	mode     BalanceMode
}

// NewPool creates a new backend pool
// Accepts a list of backend options
func newPool(backends []Backend, mode BalanceMode) *Pool {
	pool := &Pool{
		backends: make(Backends, len(backends)),
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
		backend = p.backends.MinUp(func(b *Backend) int64 { return b.Connections() })
	case ModeFirstUp:
		backend = p.backends.FirstUp()
	case ModeMinLatency:
		backend = p.backends.MinUp(func(b *Backend) int64 { return int64(b.Latency()) })
	case ModeRandom:
		backend = p.backends.Up().Random()
	case ModeWeightedLatency:
		backend = p.backends.Up().WeightedRandom(func(b *Backend) int64 {
			factor := int64(b.Latency())
			return factor * factor
		})
	}

	// Fall back on random backend
	if backend == nil {
		backend = p.backends.Random()
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

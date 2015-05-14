package balancer

import (
	"sync/atomic"
	"time"

	"gopkg.in/redis.v2"
)

type BalanceMode int

const (
	// LeastConn picks the backend with the fewest connections.
	ModeLeastConn BalanceMode = iota
	// FirstUp always picks the first available backend.
	ModeFirstUp
	// ModeMinLatency always picks the backend with the minimal latency.
	ModeMinLatency
	// ModeRandom selects backends randomly.
	ModeRandom
	// ModeWeightedLatency uses latency as a weight for random selection.
	ModeWeightedLatency
	// ModeRoundRobin round-robins across available backends.
	ModeRoundRobin
)

const minCheckInterval = 100 * time.Millisecond

// Balancer client
type Balancer struct {
	selector pool
	mode     BalanceMode
	cursor   int32
}

// New initializes a new redis balancer
func New(opts []Options, mode BalanceMode) *Balancer {
	if len(opts) == 0 {
		opts = []Options{
			{Options: redis.Options{Network: "tcp", Addr: "127.0.0.1:6379"}},
		}
	}

	balancer := &Balancer{
		selector: make(pool, len(opts)),
		mode:     mode,
	}
	for i, opt := range opts {
		balancer.selector[i] = newRedisBackend(&opt)
	}
	return balancer
}

// Next returns the next available redis client
func (b *Balancer) Next() *redis.Client { return b.pickNext().client }

// Close closes all connecitons in the balancer
func (b *Balancer) Close() (err error) {
	for _, b := range b.selector {
		if e := b.Close(); e != nil {
			err = e
		}
	}
	return
}

// Pick the next backend
func (b *Balancer) pickNext() (backend *redisBackend) {
	switch b.mode {
	case ModeLeastConn:
		backend = b.selector.MinUp(func(b *redisBackend) int64 {
			return b.Connections()
		})
	case ModeFirstUp:
		backend = b.selector.FirstUp()
	case ModeMinLatency:
		backend = b.selector.MinUp(func(b *redisBackend) int64 {
			return int64(b.Latency())
		})
	case ModeRandom:
		backend = b.selector.Up().Random()
	case ModeWeightedLatency:
		backend = b.selector.Up().WeightedRandom(func(b *redisBackend) int64 {
			factor := int64(b.Latency())
			return factor * factor
		})
	case ModeRoundRobin:
		next := int(atomic.AddInt32(&b.cursor, 1))
		backend = b.selector.Up().At(next)
	}

	// Fall back on random backend
	if backend == nil {
		backend = b.selector.Random()
	}

	// Increment the number of connections
	backend.incConnections(1)
	return
}

// --------------------------------------------------------------------

// Custom balancer options
type Options struct {
	redis.Options

	// Check interval, min 100ms, defaults to 1s
	CheckInterval time.Duration

	// Rise and Fall indicate the number of checks required to
	// mark the instance as up or down, defaults to 1
	Rise, Fall int
}

func (o *Options) getCheckInterval() time.Duration {
	if o.CheckInterval == 0 {
		return time.Second
	} else if o.CheckInterval < minCheckInterval {
		return minCheckInterval
	}
	return o.CheckInterval
}

func (o *Options) getRise() int {
	if o.Rise < 1 {
		return 1
	}
	return o.Rise
}

func (o *Options) getFall() int {
	if o.Fall < 1 {
		return 1
	}
	return o.Fall
}

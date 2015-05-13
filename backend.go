package balancer

import (
	"regexp"
	"strconv"
	"sync/atomic"
	"time"

	"gopkg.in/redis.v2"
	"gopkg.in/tomb.v2"
)

var pattern = regexp.MustCompile(`connected_clients:(\d+)`)

// Redis backend
type redisBackend struct {
	client *redis.Client
	opt    *Options

	up, successes, failures int32
	connections, latency    int64

	closer tomb.Tomb
}

func newRedisBackend(opt *Options) *redisBackend {
	backend := &redisBackend{
		client: redis.NewClient(&opt.Options),
		opt:    opt,
		up:     1,

		connections: 1e6,
		latency:     int64(time.Minute),
	}
	backend.startLoop()
	return backend
}

// Up returns true if up
func (b *redisBackend) Up() bool { return atomic.LoadInt32(&b.up) > 0 }

// Down returns true if down
func (b *redisBackend) Down() bool { return !b.Up() }

// Connections returns the number of connections
func (b *redisBackend) Connections() int64 { return atomic.LoadInt64(&b.connections) }

// Latency returns the current latency
func (b *redisBackend) Latency() time.Duration { return time.Duration(atomic.LoadInt64(&b.latency)) }

// Close shuts down the backend
func (b *redisBackend) Close() error {
	b.closer.Kill(nil)
	return b.closer.Wait()
}

func (b *redisBackend) ping() {
	start := time.Now()
	info, err := b.client.Info().Result()
	if err != nil {
		b.updateStatus(false)
		return
	}
	atomic.StoreInt64(&b.latency, int64(time.Now().Sub(start)))

	connval := pattern.FindStringSubmatch(info)
	if len(connval) != 2 {
		b.updateStatus(false)
		return
	}
	numconns, _ := strconv.ParseInt(connval[1], 10, 64)
	atomic.StoreInt64(&b.connections, numconns)

	b.updateStatus(true)
}

func (b *redisBackend) incConnections(n int64) {
	atomic.AddInt64(&b.connections, n)
}

func (b *redisBackend) updateStatus(success bool) {
	if success {
		atomic.StoreInt32(&b.failures, 0)
		rise := b.opt.getRise()

		if n := int(atomic.AddInt32(&b.successes, 1)); n > rise {
			atomic.AddInt32(&b.successes, -1)
		} else if n == rise {
			atomic.CompareAndSwapInt32(&b.up, 0, 1)
		}
	} else {
		atomic.StoreInt32(&b.successes, 0)
		fall := b.opt.getFall()

		if n := int(atomic.AddInt32(&b.failures, 1)); n > fall {
			atomic.AddInt32(&b.failures, -1)
		} else if n == fall {
			atomic.CompareAndSwapInt32(&b.up, 1, 0)
		}
	}
}

func (b *redisBackend) startLoop() {
	interval := b.opt.getCheckInterval()
	b.ping()

	b.closer.Go(func() error {
		for {
			select {
			case <-b.closer.Dying():
				return b.client.Close()
			case <-time.After(interval):
				b.ping()
			}
		}
		return nil
	})
}

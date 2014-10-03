package balancer

import (
	"regexp"
	"strconv"
	"sync"
	"time"

	"gopkg.in/redis.v2"
	"gopkg.in/tomb.v2"
)

var pattern = regexp.MustCompile(`connected_clients:(\d+)`)

type Backend struct {
	// TCP bind address or unix socket path, defaults to 127.0.0.1:6379
	Addr string

	// Network type, either "tcp" or "unix", defaults to TCP
	Network string

	// Check interval, min 100ms, defaults to 1s
	CheckInterval time.Duration

	up          bool
	connections int64
	latency     time.Duration

	closer tomb.Tomb
	mutex  sync.Mutex
}

// Up returns true if up
func (b *Backend) Up() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.up
}

// Down returns true if down
func (b *Backend) Down() bool {
	return !b.Up()
}

// Connections returns the number of connections
func (b *Backend) Connections() int64 {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.connections
}

// Latency returns the current latency
func (b *Backend) Latency() time.Duration {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.latency
}

func (b *Backend) normalize() *Backend {
	clone := new(Backend)
	*clone = *b

	if clone.Addr == "" {
		clone.Addr = "127.0.0.1:6379"
	}
	if clone.Network != "unix" {
		clone.Network = "tcp"
	}
	if clone.CheckInterval < 100*time.Millisecond {
		clone.CheckInterval = time.Second
	}
	return clone
}

func (b *Backend) newClient() *redis.Client {
	if b.Network == "unix" {
		return redis.NewUnixClient(&redis.Options{Addr: b.Addr})
	}
	return redis.NewTCPClient(&redis.Options{Addr: b.Addr})
}

func (b *Backend) poll() {
	client := b.newClient()
	defer client.Close()

	start := time.Now()
	info, err := client.Info().Result()
	latency := time.Now().Sub(start)
	if err != nil {
		b.set(false, 0, latency)
		return
	}

	cres := pattern.FindStringSubmatch(info)
	if len(cres) != 2 {
		b.set(false, 0, latency)
		return
	}

	cnum, _ := strconv.ParseInt(cres[1], 10, 64)
	b.set(true, cnum, latency)
}

func (b *Backend) set(up bool, conns int64, latency time.Duration) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.up = up
	b.connections = conns
	b.latency = latency
}

func (b *Backend) inc() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.connections++
}

func (b *Backend) start() {
	b.poll()

	b.closer.Go(func() error {
		for {
			select {
			case <-b.closer.Dying():
				return nil
			case <-time.After(b.CheckInterval):
				// continue
			}
			b.poll()
		}
		return nil
	})
}

func (b *Backend) close() error {
	b.closer.Kill(nil)
	return b.closer.Wait()
}

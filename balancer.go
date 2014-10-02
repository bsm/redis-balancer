package balancer

import (
	"errors"
	"net"
	"time"

	"gopkg.in/redis.v2"
)

type BalanceMode int

const (
	// LeastConn picks the backend with the fewest connections.
	ModeLeastConn BalanceMode = iota
	// FirstUp always picks the first available backend.
	ModeFirstUp
	// ModeLatency always picks the backend with the minimal latency.
	ModeLatency
)

var (
	ErrNoBackends = errors.New("redis-balancer: no backends provided")
)

// Client
type Client struct {
	redis.Client
	pool *Pool
}

// New client initializes a new redis-client
func NewClient(backends []Backend, mode BalanceMode, opt *redis.Options) *Client {
	if len(backends) < 1 {
		backends = []Backend{Backend{}}
	}

	if opt == nil {
		opt = new(redis.Options)
	}
	if opt.DialTimeout < 1 {
		opt.DialTimeout = 5 * time.Second
	}

	pool := newPool(backends, mode)
	opt.Dialer = func() (net.Conn, error) {
		network, address := pool.Next()
		return net.DialTimeout(network, address, opt.DialTimeout)
	}

	return &Client{*redis.NewClient(opt), pool}
}

// Close closes client and underlying pool
func (c *Client) Close() error {
	c.pool.Close()
	return c.Client.Close()
}

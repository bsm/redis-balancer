package balancer

import (
	"errors"
	"net"

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
	ErrOptionsNil = errors.New("redis-balancer: options cannot be nil")
	ErrNoBackends = errors.New("redis-balancer: no backends provided")
)

// Client
type Client struct {
	redis.Client
	pool *Pool
}

// New client initializes a new redis-client
func NewClient(backends []Backend, mode BalanceMode, opt *redis.Options) (*Client, error) {
	if len(backends) < 1 {
		return nil, ErrNoBackends
	} else if opt == nil {
		return nil, ErrOptionsNil
	}

	pool := newPool(backends, mode)
	client := redis.DialClient(opt, func() (net.Conn, error) {
		network, address := pool.Next()
		return net.DialTimeout(network, address, opt.DialTimeout)
	})

	return &Client{*client, pool}, nil
}

// Close closes client and underlying pool
func (c *Client) Close() error {
	c.pool.Close()
	return c.Client.Close()
}

package main

import (
	"log"
	"time"

	"github.com/bsm/redis-balancer"
	"gopkg.in/redis.v2"
)

func main() {
	client, err := balancer.NewClient(
		[]balancer.Backend{
			balancer.Backend{Addr: "host-1:6379", CheckInterval: 600 * time.Millisecond},
			balancer.Backend{Addr: "/tmp/redis.sock", Network: "unix"},
			balancer.Backend{Addr: "host-2:6379", CheckInterval: 800 * time.Millisecond},
			balancer.Backend{Addr: "host-2:6380"},
		},
		balancer.ModeLeastConn,
		&redis.Options{DialTimeout: time.Second},
	)
	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(res)
}

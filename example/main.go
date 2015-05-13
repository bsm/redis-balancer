package main

import (
	"log"
	"time"

	"github.com/bsm/redis-balancer"
	"gopkg.in/redis.v2"
)

func main() {
	clients := balancer.New(
		[]balancer.Options{
			{Options: redis.Options{Network: "tcp", Addr: "host-1:6379"}, CheckInterval: 600 * time.Millisecond},
			{Options: redis.Options{Network: "unix", Addr: "/tmp/redis.sock"}},
			{Options: redis.Options{Network: "tcp", Addr: "host-2:6379"}, CheckInterval: 800 * time.Millisecond},
			{Options: redis.Options{Network: "tcp", Addr: "host-2:6380"}},
		},
		balancer.ModeLeastConn,
	)
	defer clients.Close()

	client := clients.Next()
	res, err := client.Ping().Result()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(res)
}

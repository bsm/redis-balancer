package balancer

import (
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backends", func() {
	var subject Backends
	var addrsOf = func(be Backends) []string {
		addrs := make([]string, len(be))
		for i, b := range be {
			addrs[i] = b.Addr
		}
		return addrs
	}

	BeforeEach(func() {
		rand.Seed(100)
		subject = Backends{
			&Backend{Network: "tcp", Addr: "host-1:6379", up: false, connections: 4, latency: time.Millisecond},
			&Backend{Network: "tcp", Addr: "host-2:6379", up: true, connections: 12, latency: 2 * time.Millisecond},
			&Backend{Network: "tcp", Addr: "host-3:6379", up: true, connections: 8, latency: 3 * time.Millisecond},
			&Backend{Network: "tcp", Addr: "host-4:6379", up: true, connections: 16, latency: 1 * time.Millisecond},
		}
	})

	It("should select up", func() {
		Expect(addrsOf(subject.Up())).To(Equal([]string{
			"host-2:6379",
			"host-3:6379",
			"host-4:6379",
		}))
	})

	It("should select first up", func() {
		Expect(Backends{}.FirstUp()).To(BeNil())
		Expect(subject.FirstUp().Addr).To(Equal("host-2:6379"))
	})

	It("should select min up", func() {
		Expect(Backends{}.MinUp(func(b *Backend) int64 { return 100 })).To(BeNil())
		Expect(subject.MinUp(func(b *Backend) int64 { return b.Connections() }).Addr).To(Equal("host-3:6379"))
		Expect(subject.MinUp(func(b *Backend) int64 { return int64(b.Latency()) }).Addr).To(Equal("host-4:6379"))
	})

	It("should select random", func() {
		res := make(map[string]int)
		for i := 0; i < 1000; i++ {
			res[subject.Random().Addr]++
		}
		Expect(res).To(Equal(map[string]int{"host-1:6379": 259, "host-2:6379": 241, "host-3:6379": 265, "host-4:6379": 235}))
	})

	It("should select weighted-random", func() {
		Expect(Backends{}.WeightedRandom(func(b *Backend) int64 { return 100 })).To(BeNil())

		res := make(map[string]int)
		for i := 0; i < 1000; i++ {
			res[subject.WeightedRandom(func(b *Backend) int64 { return b.Connections() }).Addr]++
		}
		Expect(res).To(Equal(map[string]int{"host-1:6379": 418, "host-2:6379": 204, "host-3:6379": 302, "host-4:6379": 76}))
	})

})

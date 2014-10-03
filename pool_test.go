package balancer

import (
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pool", func() {
	var subject *Pool
	var nextAddr = func() string {
		_, addr := subject.Next()
		return addr
	}

	BeforeEach(func() {
		rand.Seed(100)
		subject = &Pool{backends: []*Backend{
			&Backend{Network: "tcp", Addr: "host-1:6379", up: false, connections: 0, latency: time.Millisecond},
			&Backend{Network: "tcp", Addr: "host-2:6379", up: true, connections: 10, latency: 2 * time.Millisecond},
			&Backend{Network: "tcp", Addr: "host-3:6379", up: true, connections: 8, latency: 3 * time.Millisecond},
			&Backend{Network: "tcp", Addr: "host-4:6379", up: true, connections: 14, latency: 1 * time.Millisecond},
		}, mode: ModeFirstUp}
	})

	It("should pick next backend (first-up)", func() {
		subject.mode = ModeFirstUp
		Expect(nextAddr()).To(Equal("host-2:6379"))
		Expect(nextAddr()).To(Equal("host-2:6379"))
		Expect(nextAddr()).To(Equal("host-2:6379"))
		Expect(nextAddr()).To(Equal("host-2:6379"))
		Expect(subject.backends[1].connections).To(Equal(int64(14)))
	})

	It("should pick next backend (least-conn)", func() {
		subject.mode = ModeLeastConn
		Expect(nextAddr()).To(Equal("host-3:6379"))
		Expect(nextAddr()).To(Equal("host-3:6379"))
		Expect(nextAddr()).To(Equal("host-2:6379"))
		Expect(nextAddr()).To(Equal("host-3:6379"))
		Expect(nextAddr()).To(Equal("host-2:6379"))
		Expect(subject.backends[1].connections).To(Equal(int64(12)))
		Expect(subject.backends[2].connections).To(Equal(int64(11)))
	})

	It("should pick next backend (min-latency)", func() {
		subject.mode = ModeMinLatency
		Expect(nextAddr()).To(Equal("host-4:6379"))
		Expect(nextAddr()).To(Equal("host-4:6379"))
		Expect(nextAddr()).To(Equal("host-4:6379"))
		Expect(nextAddr()).To(Equal("host-4:6379"))
		Expect(subject.backends[3].connections).To(Equal(int64(18)))
	})

	It("should pick next backend (randomly)", func() {
		subject.mode = ModeRandom
		Expect(nextAddr()).To(Equal("host-3:6379"))
		Expect(nextAddr()).To(Equal("host-4:6379"))
		Expect(nextAddr()).To(Equal("host-3:6379"))
		Expect(nextAddr()).To(Equal("host-2:6379"))
		Expect(subject.backends[3].connections).To(Equal(int64(15)))
	})

	It("should pick next backend (weighted-latency)", func() {
		subject.mode = ModeWeightedLatency
		Expect(nextAddr()).To(Equal("host-4:6379"))
		Expect(nextAddr()).To(Equal("host-4:6379"))
		Expect(nextAddr()).To(Equal("host-2:6379"))
		Expect(nextAddr()).To(Equal("host-2:6379"))
		Expect(nextAddr()).To(Equal("host-4:6379"))
		Expect(subject.backends[1].connections).To(Equal(int64(12)))
		Expect(subject.backends[3].connections).To(Equal(int64(17)))
	})

	It("should fallback on random when everything down", func() {
		subject.backends[1].up = false
		subject.backends[2].up = false
		subject.backends[3].up = false

		Expect(nextAddr()).To(Equal("host-4:6379"))
		Expect(nextAddr()).To(Equal("host-1:6379"))
		Expect(nextAddr()).To(Equal("host-1:6379"))
		Expect(nextAddr()).To(Equal("host-1:6379"))
		Expect(nextAddr()).To(Equal("host-3:6379"))
	})

})

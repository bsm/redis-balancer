package balancer

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backend", func() {
	var subject *Backend
	var original = &Backend{Addr: "localhost:6379", CheckInterval: 2 * time.Second}

	BeforeEach(func() {
		subject = original.normalize()
	})

	It("should normalize", func() {
		Expect(subject).To(Equal(&Backend{
			Addr:          "localhost:6379",
			Network:       "tcp",
			CheckInterval: 2 * time.Second,
		}))
		Expect(new(Backend).normalize()).To(Equal(&Backend{
			Addr:          "127.0.0.1:6379",
			Network:       "tcp",
			CheckInterval: time.Second,
		}))
	})

	It("should not mutate original", func() {
		subject.CheckInterval = 3 * time.Second
		Expect(original.CheckInterval).To(Equal(2 * time.Second))
	})

	It("should start, check and stop", func() {
		subject.start()
		defer subject.close()

		Expect(subject.Up()).To(BeTrue())
		Expect(subject.Connections()).To(BeNumerically(">", 0))
		Expect(subject.Latency()).To(BeNumerically(">", 0))
	})

})

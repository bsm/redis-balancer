package balancer

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var subject *Client

	BeforeEach(func() {
		subject = NewClient(nil, ModeFirstUp, nil)
	})

	AfterEach(func() {
		subject.Close()
	})

	It("should initialize with defaults", func() {
		Expect(subject.pool.backends).To(HaveLen(1))
		Expect(subject.pool.backends[0].Addr).To(Equal("127.0.0.1:6379"))
	})

})

/*************************************************************************
 * GINKGO TEST HOOK
 *************************************************************************/

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "redis-balancer")
}

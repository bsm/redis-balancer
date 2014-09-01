package balancer

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

/*************************************************************************
 * GINKGO TEST HOOK
 *************************************************************************/

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "redis-balancer")
}

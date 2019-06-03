package types_test

import (
	"github.com/cloudfoundry-community/bosh-vault/store"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"

	"testing"
)

var healthySimpleStore store.SimpleStore

func TestTypes(t *testing.T) {
	RegisterFailHandler(Fail)

	var err error
	var listener1 net.Listener
	healthySimpleStore, listener1, err = store.TestHealthySimpleStore(t)
	Expect(err).NotTo(HaveOccurred())

	RunSpecs(t, "Types Suite")

	listener1.Close()
}

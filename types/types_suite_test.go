package types_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/zipcar/bosh-vault/store"
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

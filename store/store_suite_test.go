package store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/zipcar/bosh-vault/logger"
	"github.com/zipcar/bosh-vault/store"
	"io/ioutil"
	"net"
	"testing"
)

var healthySimpleStore store.SimpleStore
var uninitializedVaultSimpleStore store.SimpleStore
var sealedVaultSimpleStore store.SimpleStore
var storeListeners []net.Listener

func TestStore(t *testing.T) {
	RegisterFailHandler(Fail)
	// Make sure logger singleton is available
	logger.Log = logrus.New()
	logger.Log.Out = ioutil.Discard

	// doing the store setup here is best because it swallows the output from the in memeory vault stores that it spawns
	setupStores(t)

	RunSpecs(t, "Store Suite")
}

// Not in before suite because of the in memory vault store instantiation flooding output in test results
func setupStores(t *testing.T) {
	var err error

	var listener1 net.Listener
	healthySimpleStore, listener1, err = store.TestHealthySimpleStore(t)
	Expect(err).NotTo(HaveOccurred())
	storeListeners = append(storeListeners, listener1)

	var listener2 net.Listener
	sealedVaultSimpleStore, listener2, err = store.TestSealedSimpleStore(t)
	Expect(err).NotTo(HaveOccurred())
	storeListeners = append(storeListeners, listener2)

	var listener3 net.Listener
	uninitializedVaultSimpleStore, listener3, err = store.TestUninitializedVaultSimpleStore(t)
	Expect(err).NotTo(HaveOccurred())
	storeListeners = append(storeListeners, listener3)
}

var _ = AfterSuite(func() {
	for _, listener := range storeListeners {
		defer listener.Close()
	}
})

package master

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMaster_getMasterPeers(t *testing.T) {
	m := &Master{
		replicas:  3,
		namespace: "test-namespace",
		port:      9000,
	}

	peers := m.getMasterPeers()
	assert.Equal(t, "0:master-0.master-service.test-namespace.svc.cluster.local:9000,1:master-1.master-service.test-namespace.svc.cluster.local:9000,2:master-2.master-service.test-namespace.svc.cluster.local:9000", peers)
}

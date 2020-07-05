package cluster

import (
	"github.com/rook/rook/pkg/clusterd"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type cluster struct {
	stopCh chan struct{}
}

// Add creates a new ChubaoCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, context *clusterd.Context) error {
	return add(mgr, newReconciler(mgr, context))
}

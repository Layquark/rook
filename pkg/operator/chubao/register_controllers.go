package chubao

import (
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/operator/chubao/cluster"
	"github.com/rook/rook/pkg/operator/chubao/console"
	"github.com/rook/rook/pkg/operator/chubao/objectstore"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManagerFuncs is a list of functions to add all Controllers to the Manager
var AddToManagerFuncs []func(m manager.Manager, context *clusterd.Context) error

func init() {
	AddToManagerFuncs = append(AddToManagerFuncs, console.Add)
	AddToManagerFuncs = append(AddToManagerFuncs, objectstore.Add)
}

// AddToManager adds all Controllers to the Manager
func AddToManager(m manager.Manager, context *clusterd.Context) error {
	// cluster as base
	if err := cluster.Add(m, context); err != nil {
		return err
	}

	for _, f := range AddToManagerFuncs {
		if err := f(m, context); err != nil {
			return err
		}
	}
	return nil
}

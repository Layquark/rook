/*
Copyright 2019 The Rook Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"github.com/pkg/errors"
	"github.com/rook/rook/pkg/clusterd"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// AddToManager adds all the registered controllers to the passed manager.
// each controller package will have an Add method listed in AddToManagerFuncs
// which will setup all the necessary watch
func AddToManager(m manager.Manager, context *clusterd.Context, clusterController *ClusterController) error {
	if context == nil {
		return errors.New("nil context passed")
	}

	// Run ChubaoFSCluster CR
	if err := Add(m, context, clusterController); err != nil {
		return err
	}

	//// Add Ceph child CR controllers
	//for _, f := range AddToManagerFuncs {
	//	if err := f(m, c.ClusterdContext); err != nil {
	//		return err
	//	}
	//}
	//
	//// Add maintenance controllers
	//for _, f := range AddToManagerFuncsMaintenance {
	//	if err := f(m, c); err != nil {
	//		return err
	//	}
	//}
	//
	//// If machine disruption budget is enabled let's add the controllers
	//if EnableMachineDisruptionBudget {
	//	for _, f := range MachineDisruptionBudgetAddToManagerFuncs {
	//		if err := f(m, c); err != nil {
	//			return err
	//		}
	//	}
	//}

	return nil
}

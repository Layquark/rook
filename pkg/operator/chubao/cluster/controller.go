/*
Copyright 2016 The Rook Authors. All rights reserved.

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

// Package cluster to manage a Chubao cluster.
package cluster

import (
	"github.com/coreos/pkg/capnslog"
	"github.com/rook/rook/pkg/clusterd"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sync"
)

const (
	controllerName = "chubao-cluster-controller"
)

var (
	logger = capnslog.NewPackageLogger("github.com/rook/rook", controllerName)
)

// List of object resources to watch by the controller
var objectsToWatch = []runtime.Object{
	&appsv1.Deployment{TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: appsv1.SchemeGroupVersion.String()}},
	&corev1.Service{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: corev1.SchemeGroupVersion.String()}},
	&corev1.Secret{TypeMeta: metav1.TypeMeta{Kind: "Secret", APIVersion: corev1.SchemeGroupVersion.String()}},
	&corev1.ConfigMap{TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: corev1.SchemeGroupVersion.String()}},
}

// ControllerTypeMeta Sets the type meta for the controller main object
//var ControllerTypeMeta = metav1.TypeMeta{
//	Kind:       opcontroller.ClusterResource.Kind,
//	APIVersion: opcontroller.ClusterResource.APIVersion,
//}

// ClusterController controls an instance of a Rook cluster
type ClusterController struct {
	context                 *clusterd.Context
	operatorConfigCallbacks []func() error
	addClusterCallbacks     []func() error
	csiConfigMutex          *sync.Mutex
	nodeStore               cache.Store
	//osdChecker              *osd.OSDHealthMonitor
	client         client.Client
	namespacedName types.NamespacedName
	clusterMap     map[string]*cluster
}

// ReconcileChubaoFSCluster reconciles a CephFilesystem object
type ReconcileChubaoFSCluster struct {
	client            client.Client
	scheme            *runtime.Scheme
	context           *clusterd.Context
	clusterController *ClusterController
}

// Add creates a new CephCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, context *clusterd.Context, clusterController *ClusterController) error {
	return add(mgr, newReconciler(mgr, context, clusterController), context)
}

func newReconciler(mgr manager.Manager, context *clusterd.Context, clusterController *ClusterController) reconcile.Reconciler {
	return &ReconcileChubaoFSCluster{
		client:            mgr.GetClient(),
		scheme:            mgr.GetScheme(),
		context:           context,
		clusterController: clusterController,
	}
}

func add(mgr manager.Manager, r reconcile.Reconciler, context *clusterd.Context) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	logger.Info("successfully started")

	// Watch for changes on the CephCluster CR object
	err = c.Watch(
		&source.Kind{
			//Type: &chubao_v1alpha1.ChubaoFSCluster{
			//	TypeMeta: ControllerTypeMeta,
			//},
		},
		&handler.EnqueueRequestForObject{},
	)
	if err != nil {
		return err
	}

	return nil
}

// Reconcile reads that state of the cluster for a CephCluster object and makes changes based on the state read
// and what is in the cephCluster.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileChubaoFSCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// workaround because the rook logging mechanism is not compatible with the controller-runtime loggin interface
	reconcileResponse, err := r.reconcile(request)
	if err != nil {
		logger.Errorf("failed to reconcile. %v", err)
	}

	return reconcileResponse, err
}

func (r *ReconcileChubaoFSCluster) reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Pass the client context to the ClusterController
	r.clusterController.client = r.client

	// Used by functions not part of the ClusterController struct but are given the context to execute actions
	r.clusterController.context.Client = r.client

	// Pass object name and namespace
	r.clusterController.namespacedName = request.NamespacedName
	//

	// Return and do not requeue
	return reconcile.Result{}, nil
}

// NewClusterController create controller for watching cluster custom resources created
func NewClusterController(context *clusterd.Context, operatorConfigCallbacks []func() error, addClusterCallbacks []func() error) *ClusterController {
	return &ClusterController{
		context:                 context,
		operatorConfigCallbacks: operatorConfigCallbacks,
		addClusterCallbacks:     addClusterCallbacks,
		csiConfigMutex:          &sync.Mutex{},
		clusterMap:              make(map[string]*cluster),
	}
}

// StopWatch stop watchers
func (c *ClusterController) StopWatch() {
	for _, cluster := range c.clusterMap {
		close(cluster.stopCh)
	}
	c.clusterMap = make(map[string]*cluster)
}

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
	"context"
	"github.com/coreos/pkg/capnslog"
	"github.com/pkg/errors"
	chubaoapi "github.com/rook/rook/pkg/apis/chubao.rook.io/v1alpha1"
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/operator/chubao/master"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	controllerName = "chubao-cluster-controller"
)

var (
	logger = capnslog.NewPackageLogger("github.com/rook/rook", controllerName)
)

// List of object resources to watch by the controller
var ObjectsToWatch = []runtime.Object{
	&appsv1.StatefulSet{TypeMeta: metav1.TypeMeta{Kind: "StatefulSet", APIVersion: appsv1.SchemeGroupVersion.String()}},
	&appsv1.DaemonSet{TypeMeta: metav1.TypeMeta{Kind: "DaemonSet", APIVersion: appsv1.SchemeGroupVersion.String()}},
	&appsv1.Deployment{TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: appsv1.SchemeGroupVersion.String()}},
	&corev1.Service{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: corev1.SchemeGroupVersion.String()}},
	&corev1.Secret{TypeMeta: metav1.TypeMeta{Kind: "Secret", APIVersion: corev1.SchemeGroupVersion.String()}},
	&corev1.ConfigMap{TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: corev1.SchemeGroupVersion.String()}},
}

// ControllerTypeMeta Sets the type meta for the controller main object
var NodeTypeMeta = metav1.TypeMeta{
	Kind:       "Node",
	APIVersion: corev1.SchemeGroupVersion.String(),
}

// ControllerTypeMeta Sets the type meta for the controller main object
var ControllerTypeMeta = metav1.TypeMeta{
	Kind:       reflect.TypeOf(chubaoapi.ChubaoCluster{}).Name(),
	APIVersion: chubaoapi.Version,
}

// ReconcileChubaoCluster reconciles a ChubaoFS cluster
type ReconcileChubaoCluster struct {
	client  client.Client
	scheme  *runtime.Scheme
	context *clusterd.Context
}

func newReconciler(mgr manager.Manager, context *clusterd.Context) reconcile.Reconciler {
	return &ReconcileChubaoCluster{
		client:  mgr.GetClient(),
		scheme:  mgr.GetScheme(),
		context: context,
	}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(controllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	logger.Info("successfully started")

	// Watch for changes on the ChubaoCluster CR object
	err = c.Watch(
		&source.Kind{Type: &chubaoapi.ChubaoCluster{TypeMeta: ControllerTypeMeta}},
		&handler.EnqueueRequestForObject{},
	)
	if err != nil {
		return err
	}

	// Watch all other resources of the Chubao Cluster
	for _, t := range ObjectsToWatch {
		err = c.Watch(
			&source.Kind{Type: t},
			&handler.EnqueueRequestForOwner{IsController: true, OwnerType: &chubaoapi.ChubaoCluster{}},
		)
		if err != nil {
			return err
		}
	}

	return err
}

func (r *ReconcileChubaoCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger.Infof("Reconciling ChubaoCluster:%s", request.String())
	cluster := &chubaoapi.ChubaoCluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cluster)
	if err != nil {
		if kerrors.IsNotFound(err) {
			logger.Debug("ChubaoCluster resource not found. Ignoring since object must be deleted.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, errors.Wrap(err, "failed to get ChubaoCluster")
	}

	// Define a new Pod object
	pod := master.NewMasterForCR(cluster)

	// Set Memcached instance as the owner and controller
	if err := controllerutil.SetControllerReference(cluster, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)

	// Pod already exists - don't requeue
	logger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

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

// Package operator to manage Kubernetes storage.
package chubao

import (
	chubaoapi "github.com/rook/rook/pkg/apis/chubao.rook.io/v1alpha1"
	"os"
	"os/signal"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"syscall"

	"github.com/coreos/pkg/capnslog"
	"github.com/pkg/errors"
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/operator/chubao/cluster"
	"github.com/rook/rook/pkg/operator/k8sutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	logger                   = capnslog.NewPackageLogger("github.com/rook/rook", "operator")
	RookCurrentNamespaceOnly = "ROOK_CURRENT_NAMESPACE_ONLY"
)

// Operator type for managing storage
type Operator struct {
	context           *clusterd.Context
	resources         []k8sutil.CustomResource
	operatorNamespace string
}

// New creates an operator instance
func New(context *clusterd.Context, operatorNamespace string) *Operator {
	return &Operator{
		context:           context,
		resources:         []k8sutil.CustomResource{cluster.ChubaoClusterResource},
		operatorNamespace: operatorNamespace,
	}
}

func (o *Operator) cleanup(stopCh chan struct{}) {
	close(stopCh)
}

// Run the operator instance
func (o *Operator) Run() error {
	// Initialize signal handler
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the controller-runtime Manager.
	stopChan := make(chan struct{})
	mgrErrorChan := make(chan error)
	go o.startManager(o.getNamespaceForWatch(), stopChan, mgrErrorChan)

	// Signal handler to stop the operator
	for {
		select {
		case <-signalChan:
			logger.Info("shutdown signal received, exiting...")
			o.cleanup(stopChan)
			return nil
		case err := <-mgrErrorChan:
			logger.Errorf("gave up to run the operator. %v", err)
			o.cleanup(stopChan)
			return err
		}
	}
}

func (o *Operator) getNamespaceForWatch() string {
	var namespaceToWatch string
	if os.Getenv(RookCurrentNamespaceOnly) == "true" {
		logger.Infof("watching the current namespace for a chubao cluster CR")
		// POD_NAMESPACE
		namespaceToWatch = o.operatorNamespace
	} else {
		logger.Infof("watching all namespaces for chubao cluster CRs")
		namespaceToWatch = v1.NamespaceAll
	}
	return namespaceToWatch
}

func (o *Operator) startManager(namespaceToWatch string, stopCh <-chan struct{}, mgrErrorCh chan error) {
	// Set up a manager
	mgrOpts := manager.Options{
		LeaderElection: false,
		Namespace:      namespaceToWatch,
	}

	logger.Info("setting up the controller-runtime manager")
	// Create a new manager to provide shared dependencies and start components
	mgr, err := manager.New(o.context.KubeConfig, mgrOpts)
	if err != nil {
		mgrErrorCh <- errors.Wrap(err, "failed to set up overall controller-runtime manager")
		return
	}

	// Setup Scheme for all resources
	if err = chubaoapi.AddToScheme(mgr.GetScheme()); err != nil {
		mgrErrorCh <- errors.Wrap(err, "failed to set up overall controller-runtime manager")
		return
	}

	// Setup all Controllers
	if err = AddToManager(mgr, o.context); err != nil {
		mgrErrorCh <- errors.Wrap(err, "failed to add controllers to controller-runtime manager")
		return
	}

	logger.Info("starting the controller-runtime manager")
	if err := mgr.Start(stopCh); err != nil {
		mgrErrorCh <- errors.Wrap(err, "unable to run the controller-runtime manager")
		return
	}
}

// getDeploymentOwnerReference returns an OwnerReference to the rook-chubao-operator deployment
func getDeploymentOwnerReference(clientset kubernetes.Interface, namespace string) (*metav1.OwnerReference, error) {
	var deploymentRef *metav1.OwnerReference
	podName := os.Getenv(k8sutil.PodNameEnvVar)
	pod, err := clientset.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "could not find pod %q to find deployment owner reference", podName)
	}
	for _, podOwner := range pod.OwnerReferences {
		if podOwner.Kind == "ReplicaSet" {
			replicaset, err := clientset.AppsV1().ReplicaSets(namespace).Get(podOwner.Name, metav1.GetOptions{})
			if err != nil {
				return nil, errors.Wrapf(err, "could not find replicaset %q to find deployment owner reference", podOwner.Name)
			}
			for _, replicasetOwner := range replicaset.OwnerReferences {
				if replicasetOwner.Kind == "Deployment" {
					deploymentRef = &replicasetOwner
				}
			}
		}
	}
	if deploymentRef == nil {
		return nil, errors.New("could not find owner reference for rook-chubao deployment")
	}
	return deploymentRef, nil
}

/*
Copyright 2017 The Rook Authors. All rights reserved.

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

// Package attachment to manage Kubernetes storage attach events.
package cluster

import (
	chubaoapi "github.com/rook/rook/pkg/apis/chubao.rook.io/v1alpha1"
	"reflect"

	"github.com/rook/rook/pkg/operator/k8sutil"
)

const (
	CustomResourceName       = "chubaocluster"
	CustomResourceNamePlural = "chubaoclusters"
)

// ChubaoClusterResource represents the ChubaoCluster custom resource object
var ChubaoClusterResource = k8sutil.CustomResource{
	Name:    CustomResourceName,
	Plural:  CustomResourceNamePlural,
	Group:   chubaoapi.CustomResourceGroup,
	Version: chubaoapi.Version,
	Kind:    reflect.TypeOf(chubaoapi.ChubaoCluster{}).Name(),
}

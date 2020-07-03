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
package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ***************************************************************************
// IMPORTANT FOR CODE GENERATION
// If the types in this file are updated, you will need to run
// `make codegen` to generate the new types under the client/clientset folder.
// ***************************************************************************

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ChubaoCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ChubaoFSSpec `json:"spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ChubaoClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ChubaoCluster `json:"items"`
}

type ChubaoFSSpec struct {
	CFSVersion      CFSVersionSpec `json:"cfsVersion,omitempty"`
	DataDirHostPath string         `json:"dataDirHostPath,omitempty"`
	LogDirHostPath  string         `json:"logDirHostPath,omitempty"`
	Master          MasterSpec     `json:"master"`
	MetaNode        MetaNodeSpec   `json:"metaNode"`
	DataNode        DataNodeSpec   `json:"dataNode"`
}

// VersionSpec represents the settings for the cfs-server version that Rook is orchestrating.
type CFSVersionSpec struct {
	ServerImage string `json:"serverImage,omitempty"`
	ClientImage string `json:"clientImage,omitempty"`
}

type DataNodeSpec struct {
	LogLevel      string                  `json:"logLevel,omitempty"`
	Port          int32                   `json:"port,omitempty"`
	Prof          int32                   `json:"prof,omitempty"`
	ExporterPort  int32                   `json:"exporterPort,omitempty"`
	RaftHeartbeat int32                   `json:"raftHeartbeat,omitempty"`
	RaftReplica   int32                   `json:"raftReplica,omitempty"`
	Disks         []string                `json:"disks,omitempty"`
	ZoneName      string                  `json:"zoneName,omitempty"`
	NodeSelector  v1.NodeSelector         `json:"nodeSelector,omitempty"`
	Resource      v1.ResourceRequirements `json:"resource,omitempty"`
}

type MetaNodeSpec struct {
	LogLevel      string                  `json:"logLevel,omitempty"`
	TotalMem      int64                   `json:"totalMem,omitempty"`
	Port          int32                   `json:"port,omitempty"`
	Prof          int32                   `json:"prof,omitempty"`
	ExporterPort  int32                   `json:"exporterPort,omitempty"`
	RaftHeartbeat int32                   `json:"raftHeartbeat,omitempty"`
	RaftReplica   int32                   `json:"raftReplica,omitempty"`
	ZoneName      string                  `json:"zoneName,omitempty"`
	NodeSelector  v1.NodeSelector         `json:"nodeSelector,omitempty"`
	Resource      v1.ResourceRequirements `json:"resource,omitempty"`
}

type MasterSpec struct {
	Replicas            int32                   `json:"replicas,omitempty"`
	Cluster             string                  `json:"cluster,omitempty"`
	LogLevel            string                  `json:"logLevel,omitempty"`
	RetainLogs          int32                   `json:"retainLogs,omitempty"`
	Port                int32                   `json:"port,omitempty"`
	Prof                int32                   `json:"prof,omitempty"`
	ExporterPort        int32                   `json:"exporterPort,omitempty"`
	MetaNodeReservedMem int32                   `json:"metaNodeReservedMem,omitempty"`
	NodeSelector        v1.NodeSelector         `json:"nodeSelector,omitempty"`
	Resource            v1.ResourceRequirements `json:"resource,omitempty"`
}

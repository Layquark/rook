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
	rookv1 "github.com/rook/rook/pkg/apis/rook.io/v1"
	appsv1 "k8s.io/api/apps/v1"
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

// Service is a named abstraction of software service (for example, mysql) consisting of local port
// (for example 3306) that the proxy listens on, and the selector that determines which pods
// will answer requests sent through the proxy.
type ChubaoCluster struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata"`

	// Spec defines the desired identities of Cluster in this set.
	// +optional
	Spec ClusterSpec `json:"spec"`

	// Status is the current status of Cluster in this ChubaoCluster. This data
	// may be out of date by some window of time.
	// +optional
	Status ClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ChubaoClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ChubaoCluster `json:"items"`
}

type ConditionType string

const (
	ConditionIgnored     ConditionType = "Ignored"
	ConditionConnecting  ConditionType = "Connecting"
	ConditionConnected   ConditionType = "Connected"
	ConditionProgressing ConditionType = "Progressing"
	ConditionReady       ConditionType = "Ready"
	ConditionUpdating    ConditionType = "Updating"
	ConditionFailure     ConditionType = "Failure"
	ConditionUpgrading   ConditionType = "Upgrading"
	ConditionDeleting    ConditionType = "Deleting"
)

type ClusterState string

const (
	ClusterStateCreating   ClusterState = "Creating"
	ClusterStateCreated    ClusterState = "Created"
	ClusterStateUpdating   ClusterState = "Updating"
	ClusterStateConnecting ClusterState = "Connecting"
	ClusterStateConnected  ClusterState = "Connected"
	ClusterStateError      ClusterState = "Error"
)

type CleanupPolicy string

const (
	CleanupPolicyNone             CleanupPolicy = "None"
	CleanupPolicyDeleteLog        CleanupPolicy = "DeleteLog"
	CleanupPolicyDeleteData       CleanupPolicy = "DeleteData"
	CleanupPolicyDeleteDataAndLog CleanupPolicy = "DeleteDataAndLog"
	CleanupPolicyDeleteDiskData   CleanupPolicy = "DeleteDiskData"
	CleanupPolicyDeleteAll        CleanupPolicy = "DeleteAll"
)

type ClusterStatus struct {
	State        ClusterState  `json:"state,omitempty"`
	Phase        ConditionType `json:"phase,omitempty"`
	Message      string        `json:"message,omitempty"`
	Conditions   []Condition   `json:"conditions,omitempty"`
	ChubaoStatus *ChubaoStatus `json:"chubao,omitempty"`
}

type Condition struct {
	Type               ConditionType      `json:"type,omitempty"`
	Status             v1.ConditionStatus `json:"status,omitempty"`
	Reason             string             `json:"reason,omitempty"`
	Message            string             `json:"message,omitempty"`
	LastHeartbeatTime  metav1.Time        `json:"lastHeartbeatTime,omitempty"`
	LastTransitionTime metav1.Time        `json:"lastTransitionTime,omitempty"`
}

// A ClusterSpec is the specification of a ChubaoCluster.
type ClusterSpec struct {
	// CFSVersion defines image version and strategy of pulling image
	// +optional
	CFSVersion CFSVersionSpec `json:"cfsVersion,omitempty"`

	// DataDirHostPath defines data path for the component(Master, DataNode, and MetaNode)
	// It is a host path, if unspecified, defaults to /var/lib/chubao.
	// +optional
	DataDirHostPath string `json:"dataDirHostPath"`

	// LogDirHostPath defines logs path for the component(Master, DataNode, and MetaNode)
	// It is a host path, if unspecified, defaults to /var/log/chubao.
	// +optional
	LogDirHostPath string `json:"logDirHostPath"`

	// Master component in ChubaoCluster
	// +optional
	Master MasterSpec `json:"master,omitempty"`

	// MetaNode component in ChubaoCluster
	MetaNode MetaNodeSpec `json:"metaNode,omitempty"`

	// DataNode component in ChubaoCluster
	DataNode DataNodeSpec `json:"dataNode,omitempty"`

	// Consul component in ChubaoCluster
	// +optional
	Consul ConsulSpec `json:"consul,omitempty"`

	Stroage *StorageSpec `json:"consul,omitempty"`

	// Indicates user intent when deleting a cluster; blocks orchestration and should not be set if cluster
	// deletion is not imminent.
	// +optional
	CleanupPolicy CleanupPolicy `json:"cleanupPolicy,omitempty"`
}

type StorageSpec struct {
}

type ConsulSpec struct {
	Port            int32                   `json:"port,omitempty"`
	Image           string                  `json:"image,omitempty"`
	ImagePullPolicy v1.PullPolicy           `json:"imagePullPolicy,omitempty"`
	Placement       *rookv1.Placement       `json:"placement,omitempty"`
	Resources       v1.ResourceRequirements `json:"resources,omitempty"`
}

type ChubaoStatus struct {
	Health         string                         `json:"health,omitempty"`
	Details        map[string]ChubaoHealthMessage `json:"details,omitempty"`
	LastChecked    string                         `json:"lastChecked,omitempty"`
	LastChanged    string                         `json:"lastChanged,omitempty"`
	PreviousHealth string                         `json:"previousHealth,omitempty"`
}

type ChubaoHealthMessage struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// VersionSpec represents the settings for the cfs-server version that Rook is orchestrating.
type CFSVersionSpec struct {
	// Docker image name for ChubaoFS Server.
	// More info: https://kubernetes.io/docs/concepts/containers/images
	// This field is optional to allow higher level config management to default or override
	// container images in workload controllers like Deployments and StatefulSets.
	// +optional
	ServerImage string `json:"serverImage"`

	// Image pull policy.
	// One of Always, Never, IfNotPresent.
	// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
	// Cannot be updated.
	// +optional
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy,omitempty" protobuf:"bytes,14,opt,name=imagePullPolicy,casttype=PullPolicy"`
}

type DataNodeSpec struct {
	LogLevel          string                         `json:"logLevel,omitempty"`
	Port              int32                          `json:"port,omitempty"`
	Prof              int32                          `json:"prof,omitempty"`
	ExporterPort      int32                          `json:"exporterPort,omitempty"`
	RaftHeartbeatPort int32                          `json:"raftHeartbeatPort,omitempty"`
	RaftReplicaPort   int32                          `json:"raftReplicaPort,omitempty"`
	Disks             []string                       `json:"disks"`
	Zone              string                         `json:"zone,omitempty"`
	Placement         *rookv1.Placement              `json:"placement,omitempty"`
	UpdateStrategy    appsv1.DaemonSetUpdateStrategy `json:"updateStrategy,omitempty"`
	Resource          v1.ResourceRequirements        `json:"resource,omitempty"`
}

type MetaNodeSpec struct {
	LogLevel          string                         `json:"logLevel,omitempty"`
	TotalMem          int64                          `json:"totalMem,omitempty"`
	Port              int32                          `json:"port,omitempty"`
	Prof              int32                          `json:"prof,omitempty"`
	ExporterPort      int32                          `json:"exporterPort,omitempty"`
	RaftHeartbeatPort int32                          `json:"raftHeartbeatPort,omitempty"`
	RaftReplicaPort   int32                          `json:"raftReplicaPort,omitempty"`
	Zone              string                         `json:"zone,omitempty"`
	Placement         *rookv1.Placement              `json:"placement,omitempty"`
	UpdateStrategy    appsv1.DaemonSetUpdateStrategy `json:"updateStrategy,omitempty"`
	Resource          v1.ResourceRequirements        `json:"resource,omitempty"`
}

// A MasterSpec is the specification of Master component in the ChubaoCluster.
type MasterSpec struct {
	// replicas is the desired number of replicas of the given Template.
	// These are replicas in the sense that they are instantiations of the
	// same Template, but individual replicas also have a consistent identity.
	// If unspecified, defaults to 3.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	LogLevel            string                           `json:"logLevel,omitempty"`
	RetainLogs          int32                            `json:"retainLogs,omitempty"`
	Port                int32                            `json:"port,omitempty"`
	Prof                int32                            `json:"prof,omitempty"`
	ExporterPort        int32                            `json:"exporterPort,omitempty"`
	MetaNodeReservedMem int64                            `json:"metaNodeReservedMem,omitempty"`
	Placement           *rookv1.Placement                `json:"placement,omitempty"`
	UpdateStrategy      appsv1.StatefulSetUpdateStrategy `json:"updateStrategy,omitempty"`
	Resource            v1.ResourceRequirements          `json:"resource,omitempty"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ChubaoMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              MonitorSpec   `json:"spec"`
	Status            MonitorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ChubaoMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ChubaoMonitor `json:"items"`
}

type MonitorSpec struct {
	ConsulUrl string `json:"consulUrl"`
	Password  string `json:"password"`
}

type GrafanaStatus string

const (
	GrafanaStatusReady   GrafanaStatus = "Ready"
	GrafanaStatusFailure GrafanaStatus = "Failure"
	GrafanaStatusUnknown GrafanaStatus = "Unknown"
)

type PrometheusStatus string

const (
	PrometheusStatusReady   PrometheusStatus = "Ready"
	PrometheusStatusFailure PrometheusStatus = "Failure"
	PrometheusStatusUnknown PrometheusStatus = "Unknown"
)

type MonitorStatus struct {
	Grafana    GrafanaStatus    `json:"grafana"`
	Prometheus PrometheusStatus `json:"prometheus"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ChubaoObjectStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ObjectStoreSpec   `json:"spec"`
	Status            ObjectStoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ChubaoObjectStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ChubaoMonitor `json:"items"`
}

type ObjectStoreSpec struct {
	Replicas     int32                   `json:"replicas,omitempty"`
	MasterAddr   string                  `json:"masterAddr"`
	LogLevel     string                  `json:"logLevel,omitempty"`
	Port         int32                   `json:"port,omitempty"`
	Prof         int32                   `json:"prof,omitempty"`
	ExporterPort int32                   `json:"exporterPort,omitempty"`
	Region       string                  `json:"region"`
	Domains      string                  `json:"domains,omitempty"`
	Host         string                  `json:"host,omitempty"`
	Resource     v1.ResourceRequirements `json:"resource,omitempty"`
}

type ObjectStoreStatus struct {
}

package master

import (
	"fmt"
	"github.com/pkg/errors"
	chubaoapi "github.com/rook/rook/pkg/apis/chubao.rook.io/v1alpha1"
	"github.com/rook/rook/pkg/clusterd"
	"github.com/rook/rook/pkg/operator/chubao/cluster/consul"
	"github.com/rook/rook/pkg/operator/chubao/commons"
	"github.com/rook/rook/pkg/operator/chubao/constants"
	"github.com/rook/rook/pkg/operator/k8sutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/tools/record"
	"reflect"
	"strings"
)

//var logger = capnslog.NewPackageLogger("github.com/rook/rook", "chubao-master")

const (
	instanceName               = "master"
	defaultMasterServiceName   = "master-service"
	defaultServerImage         = "chubaofs/cfs-server:0.0.1"
	defaultDataDirHostPath     = "/var/lib/chubaofs"
	defaultLogDirHostPath      = "/var/log/chubaofs"
	defaultReplicas            = 3
	defaultLogLevel            = "error"
	defaultRetainLogs          = 2000
	defaultPort                = 17110
	defaultProf                = 17120
	defaultExporterPort        = 17150
	defaultMetaNodeReservedMem = 67108864

	volumeNameForLogPath       = "pod-log-path"
	volumeNameForDataPath      = "pod-data-path"
	defaultDataPathInContainer = "/cfs/data"
	defaultLogPathInContainer  = "/cfs/logs"
)

type Master struct {
	clusterObj          *chubaoapi.ChubaoCluster
	masterObj           chubaoapi.MasterSpec
	context             *clusterd.Context
	kubeInformerFactory kubeinformers.SharedInformerFactory
	ownerRef            metav1.OwnerReference
	recorder            record.EventRecorder
	namespace           string
	serverImage         string
	imagePullPolicy     corev1.PullPolicy
	dataDirHostPath     string
	logDirHostPath      string
	replicas            int32
	logLevel            string
	retainLogs          int32
	port                int32
	prof                int32
	exporterPort        int32
	metanodeReservedMem int64
}

func New(
	context *clusterd.Context,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	recorder record.EventRecorder,
	clusterObj *chubaoapi.ChubaoCluster,
	ownerRef metav1.OwnerReference) *Master {
	spec := clusterObj.Spec
	masterObj := spec.Master
	return &Master{
		context:             context,
		kubeInformerFactory: kubeInformerFactory,
		recorder:            recorder,
		clusterObj:          clusterObj,
		ownerRef:            ownerRef,
		namespace:           clusterObj.Namespace,
		masterObj:           masterObj,
		serverImage:         commons.GetStringValue(spec.CFSVersion.ServerImage, defaultServerImage),
		imagePullPolicy:     commons.GetImagePullPolicy(spec.CFSVersion.ImagePullPolicy),
		dataDirHostPath:     commons.GetStringValue(spec.DataDirHostPath, defaultDataDirHostPath),
		logDirHostPath:      commons.GetStringValue(spec.LogDirHostPath, defaultLogDirHostPath),
		replicas:            commons.GetIntValue(masterObj.Replicas, defaultReplicas),
		logLevel:            commons.GetStringValue(masterObj.LogLevel, defaultLogLevel),
		retainLogs:          commons.GetIntValue(masterObj.RetainLogs, defaultRetainLogs),
		port:                commons.GetIntValue(masterObj.Port, defaultPort),
		prof:                commons.GetIntValue(masterObj.Prof, defaultProf),
		exporterPort:        commons.GetIntValue(masterObj.ExporterPort, defaultExporterPort),
		metanodeReservedMem: commons.GetInt64Value(masterObj.MetaNodeReservedMem, defaultMetaNodeReservedMem),
	}
}

func (m *Master) Deploy() error {
	labels := masterLabels(m.clusterObj.Name)
	clientSet := m.context.Clientset
	if _, err := k8sutil.CreateOrUpdateService(clientSet, m.namespace, m.newMasterService(labels)); err != nil {
		return errors.Wrap(err, "failed to create Service for master")
	}

	statefulSet := m.newMasterStatefulSet(labels)
	msg := fmt.Sprintf("%s/%s", statefulSet.Namespace, statefulSet.Name)
	if _, err := clientSet.AppsV1().StatefulSets(m.namespace).Create(statefulSet); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return errors.Wrap(err, fmt.Sprintf("failed to create StatefulSet for master[%s]", msg))
		}

		_, err := clientSet.AppsV1().StatefulSets(m.namespace).Update(statefulSet)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to update StatefulSet for master[%s]", msg))
		}
	}

	return nil
}

func masterLabels(clusterName string) map[string]string {
	return commons.CommonLabels(constants.ComponentMaster, clusterName)
}

func (m *Master) newMasterStatefulSet(labels map[string]string) *appsv1.StatefulSet {
	statefulSet := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       reflect.TypeOf(appsv1.StatefulSet{}).Name(),
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            instanceName,
			Namespace:       m.namespace,
			OwnerReferences: []metav1.OwnerReference{m.ownerRef},
			Labels:          labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:            &m.replicas,
			ServiceName:         defaultMasterServiceName,
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			UpdateStrategy:      m.masterObj.UpdateStrategy,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: createPodSpec(m),
			},
		},
	}

	return statefulSet
}

func createPodSpec(m *Master) corev1.PodSpec {
	privileged := true
	nodeSelector := make(map[string]string)
	nodeSelector[fmt.Sprintf("%s-%s", m.namespace, constants.ComponentMaster)] = "enabled"
	pathType := corev1.HostPathDirectoryOrCreate
	return corev1.PodSpec{
		NodeSelector: nodeSelector,
		HostNetwork:  true,
		HostPID:      true,
		DNSPolicy:    corev1.DNSClusterFirstWithHostNet,
		Containers: []corev1.Container{
			{
				Name:            "master-pod",
				Image:           m.serverImage,
				ImagePullPolicy: m.imagePullPolicy,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				Command: []string{
					"/bin/bash",
				},
				Args: []string{
					"-c",
					"set -e; /cfs/bin/start.sh master; sleep 999999999d",
				},
				Env: []corev1.EnvVar{
					{Name: "CBFS_CLUSTER_NAME", Value: m.clusterObj.Name},
					{Name: "CBFS_PORT", Value: fmt.Sprintf("%d", m.port)},
					{Name: "CBFS_PROF", Value: fmt.Sprintf("%d", m.prof)},
					{Name: "CBFS_MASTER_PEERS", Value: m.getMasterPeers()},
					{Name: "CBFS_RETAIN_LOGS", Value: fmt.Sprintf("%d", m.retainLogs)},
					{Name: "CBFS_LOG_LEVEL", Value: m.logLevel},
					{Name: "CBFS_EXPORTER_PORT", Value: fmt.Sprintf("%d", m.exporterPort)},
					{Name: "CBFS_CONSUL_ADDR", Value: consul.GetConsulUrl(m.clusterObj)},
					{Name: "CBFS_METANODE_RESERVED_MEM", Value: fmt.Sprintf("%d", m.metanodeReservedMem)},
					k8sutil.PodIPEnvVar("POD_IP"),
					k8sutil.NameEnvVar(),
				},
				Ports: []corev1.ContainerPort{
					// Port Name must be no more than 15 characters
					{Name: "port", ContainerPort: m.port, Protocol: corev1.ProtocolTCP},
					{Name: "prof", ContainerPort: m.prof, Protocol: corev1.ProtocolTCP},
					{Name: "exporter-port", ContainerPort: m.exporterPort, Protocol: corev1.ProtocolTCP},
				},
				VolumeMounts: []corev1.VolumeMount{
					{Name: volumeNameForLogPath, MountPath: defaultLogPathInContainer},
					{Name: volumeNameForDataPath, MountPath: defaultDataPathInContainer},
				},
				Resources: m.masterObj.Resource,
			},
		},
		Volumes: []corev1.Volume{
			{
				Name:         volumeNameForLogPath,
				VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: m.logDirHostPath, Type: &pathType}},
			},
			{
				Name:         volumeNameForDataPath,
				VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: m.dataDirHostPath, Type: &pathType}},
			},
		},
	}
}

func (m *Master) newMasterService(labels map[string]string) *corev1.Service {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       reflect.TypeOf(corev1.Service{}).Name(),
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            defaultMasterServiceName,
			Namespace:       m.namespace,
			OwnerReferences: []metav1.OwnerReference{m.ownerRef},
			Labels:          labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "port", Port: m.port, Protocol: corev1.ProtocolTCP,
				},
			},
			Selector: labels,
		},
	}
	return service
}

// 1:master-0.master-service.svc.cluster.local:17110,2:master-1.master-service.svc.cluster.local:17110,3:master-2.master-service.svc.cluster.local:17110
func (m *Master) getMasterPeers() string {
	urls := make([]string, 0)
	for i := 0; i < int(m.replicas); i++ {
		urls = append(urls, fmt.Sprintf("%d:%s-%d.%s.%s.%s:%d", i+1, instanceName, i,
			defaultMasterServiceName, m.namespace, constants.ServiceDomainSuffix, m.port))
	}

	return strings.Join(urls, ",")
}

// master-0.master-service.svc.cluster.local:17110,master-1.master-service.svc.cluster.local:17110,master-2.master-service.svc.cluster.local:17110
func GetMasterAddrs(clusterObj *chubaoapi.ChubaoCluster) string {
	master := clusterObj.Spec.Master
	replicas := func() int {
		if master.Replicas == 0 {
			return defaultReplicas
		} else {
			return int(master.Replicas)
		}
	}()

	port := func() int {
		if master.Port == 0 {
			return defaultPort
		} else {
			return int(master.Port)
		}
	}()

	urls := make([]string, 0)
	for i := 0; i < replicas; i++ {
		urls = append(urls, fmt.Sprintf("%s-%d.%s.%s.%s:%d", instanceName, i,
			defaultMasterServiceName, clusterObj.Namespace, constants.ServiceDomainSuffix, port))
	}

	return strings.Join(urls, ",")
}

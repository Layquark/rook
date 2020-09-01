package consul

import (
	"fmt"
	"github.com/pkg/errors"
	chubaoapi "github.com/rook/rook/pkg/apis/chubao.rook.io/v1alpha1"
	"github.com/rook/rook/pkg/clusterd"
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
)

const (
	instanceName = "consul"
	serviceName  = "consul-service"

	defaultPort  = 8500
	defaultImage = "consul:1.6.1"
)

func GetConsulUrl(clusterObj *chubaoapi.ChubaoCluster) string {
	if clusterObj == nil {
		return ""
	}

	return fmt.Sprintf("http://%s.%s.%s:%d",
		serviceName,
		clusterObj.Namespace,
		constants.ServiceDomainSuffix,
		commons.GetIntValue(clusterObj.Spec.Consul.Port, defaultPort))
}

type Consul struct {
	clusterObj          *chubaoapi.ChubaoCluster
	consulObj           chubaoapi.ConsulSpec
	context             *clusterd.Context
	kubeInformerFactory kubeinformers.SharedInformerFactory
	ownerRef            metav1.OwnerReference
	recorder            record.EventRecorder
	namespace           string
	port                int32
	image               string
	imagePullPolicy     corev1.PullPolicy
}

func New(
	context *clusterd.Context,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	recorder record.EventRecorder,
	clusterObj *chubaoapi.ChubaoCluster,
	ownerRef metav1.OwnerReference) *Consul {
	consulObj := clusterObj.Spec.Consul
	return &Consul{
		context:             context,
		kubeInformerFactory: kubeInformerFactory,
		recorder:            recorder,
		clusterObj:          clusterObj,
		consulObj:           consulObj,
		ownerRef:            ownerRef,
		namespace:           clusterObj.Namespace,
		port:                commons.GetIntValue(consulObj.Port, defaultPort),
		image:               commons.GetStringValue(consulObj.Image, defaultImage),
		imagePullPolicy:     commons.GetImagePullPolicy(consulObj.ImagePullPolicy),
	}
}

func (consul *Consul) Deploy() error {
	labels := consulLabels(consul.clusterObj.Name)
	clientset := consul.context.Clientset
	if _, err := k8sutil.CreateOrUpdateService(clientset, consul.namespace, consul.newConsulService(labels)); err != nil {
		return errors.Wrap(err, "failed to create Service for master")
	}

	deployment := consul.newConsulDeployment(labels)
	msg := fmt.Sprintf("%s/%s", deployment.Namespace, deployment.Name)
	if _, err := clientset.AppsV1().Deployments(consul.namespace).Create(deployment); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return errors.Wrap(err, fmt.Sprintf("failed to create Deployment for consul[%s]", msg))
		}

		_, err := clientset.AppsV1().Deployments(consul.namespace).Update(deployment)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to update Deployment for consul[%s]", msg))
		}
	}

	return nil
}

func consulLabels(clusterName string) map[string]string {
	return commons.CommonLabels(constants.ComponentConsul, clusterName)
}

func (consul *Consul) newConsulService(labels map[string]string) *corev1.Service {
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       reflect.TypeOf(corev1.Service{}).Name(),
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            serviceName,
			Namespace:       consul.namespace,
			OwnerReferences: []metav1.OwnerReference{consul.ownerRef},
			Labels:          labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "port",
					Port:     consul.port,
					Protocol: corev1.ProtocolTCP,
				},
			},
			Selector: labels,
		},
	}
	return service
}

func (consul *Consul) newConsulDeployment(labels map[string]string) *appsv1.Deployment {
	replicas := int32(1)
	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       reflect.TypeOf(appsv1.Deployment{}).Name(),
			APIVersion: appsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            instanceName,
			Namespace:       consul.namespace,
			OwnerReferences: []metav1.OwnerReference{consul.ownerRef},
			Labels:          labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: createPodSpec(consul),
			},
		},
	}

	return deployment
}

func createPodSpec(consul *Consul) corev1.PodSpec {
	privileged := true
	pod := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            "consul-pod",
				Image:           consul.image,
				ImagePullPolicy: consul.imagePullPolicy,
				SecurityContext: &corev1.SecurityContext{
					Privileged: &privileged,
				},
				Ports: []corev1.ContainerPort{
					{
						Name: "port", ContainerPort: consul.port, Protocol: corev1.ProtocolTCP,
					},
				},
				Resources: consul.consulObj.Resources,
			},
		},
	}

	return pod
}

package datanode

import (
	chubaoapi "github.com/rook/rook/pkg/apis/chubao.rook.io/v1alpha1"
	"github.com/rook/rook/pkg/clusterd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DataNode struct {
}

func (dn *DataNode) Start() error {
	return nil
}

func New(context *clusterd.Context, namespace string, clusterObj *chubaoapi.ChubaoCluster, ownerRef metav1.OwnerReference) *DataNode {
	return nil
}

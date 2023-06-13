package service

import (
	"context"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud/request"
)

type KubernetesExperimental interface {
	Kubernetes

	DeleteKubernetesNodeGroupNode(ctx context.Context, r *request.DeleteKubernetesNodeGroupNodeRequest) error
	GetKubernetesNodeGroupDetails(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroupDetails, error)
}

// DeleteKubernetesNodeGroupNode deletes an existing node from node group (EXPERIMENTAL).
func (s *Service) DeleteKubernetesNodeGroupNode(ctx context.Context, r *request.DeleteKubernetesNodeGroupNodeRequest) error {
	return s.delete(ctx, r)
}

// GetKubernetesNodeGroupNode retrieves details of a node in a node group (EXPERIMENTAL).
func (s *Service) GetKubernetesNodeGroupDetails(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroupDetails, error) {
	ng := upcloud.KubernetesNodeGroupDetails{}
	return &ng, s.get(ctx, r.RequestURL(), &ng)
}

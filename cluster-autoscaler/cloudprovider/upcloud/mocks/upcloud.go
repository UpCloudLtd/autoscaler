package mocks

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud/request"
)

type UpCloudService struct {
	Clusters   []upcloud.KubernetesCluster
	Plans      []upcloud.KubernetesPlan
	NodeGroups []upcloud.KubernetesNodeGroup
}

func (s *UpCloudService) GetKubernetesNodeGroups(ctx context.Context, r *request.GetKubernetesNodeGroupsRequest) ([]upcloud.KubernetesNodeGroup, error) {
	return s.NodeGroups, nil

}
func (s *UpCloudService) ModifyKubernetesNodeGroup(ctx context.Context, r *request.ModifyKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error) {
	for i := range s.NodeGroups {
		if s.NodeGroups[i].Name == r.Name {
			s.NodeGroups[i].Count = r.NodeGroup.Count
			return &s.NodeGroups[i], nil
		}
	}
	return nil, fmt.Errorf("node group not found %+v", r)
}

func (s *UpCloudService) DeleteKubernetesNodeGroupNode(ctx context.Context, r *request.DeleteKubernetesNodeGroupNodeRequest) error {
	return nil
}

func (s *UpCloudService) GetKubernetesNodeGroup(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroupDetails, error) {
	for i := range s.NodeGroups {
		if s.NodeGroups[i].Name == r.Name {
			nodes := make([]upcloud.KubernetesNode, 0)
			for z := 0; z < s.NodeGroups[i].Count; z++ {
				nodes = append(nodes, upcloud.KubernetesNode{
					UUID:  fmt.Sprintf("%s-%d", r.Name, z),
					Name:  fmt.Sprintf("%s-node-%d", r.Name, z),
					State: upcloud.KubernetesNodeStateRunning,
				})
			}
			return &upcloud.KubernetesNodeGroupDetails{
				KubernetesNodeGroup: s.NodeGroups[i],
				Nodes:               nodes,
			}, nil
		}
	}
	return nil, fmt.Errorf("node group details not found %+v", r)
}

func (s *UpCloudService) GetKubernetesCluster(ctx context.Context, r *request.GetKubernetesClusterRequest) (*upcloud.KubernetesCluster, error) {
	for i := range s.Clusters {
		if s.Clusters[i].UUID == r.UUID {
			return &s.Clusters[i], nil
		}
	}
	return nil, &upcloud.Problem{Status: http.StatusNotFound}
}

func (s *UpCloudService) GetKubernetesPlans(ctx context.Context, r *request.GetKubernetesPlansRequest) ([]upcloud.KubernetesPlan, error) {
	return s.Plans, nil
}

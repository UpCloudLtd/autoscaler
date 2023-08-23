/*
Copyright 2023 The Kubernetes Authors.

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

package mocks

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud/request"
)

// UpCloudService is mock that implements UpCloudService
type UpCloudService struct {
	Clusters   []upcloud.KubernetesCluster
	Plans      []upcloud.KubernetesPlan
	NodeGroups []upcloud.KubernetesNodeGroup
}

// GetKubernetesNodeGroups list node groups
func (s *UpCloudService) GetKubernetesNodeGroups(ctx context.Context, r *request.GetKubernetesNodeGroupsRequest) ([]upcloud.KubernetesNodeGroup, error) {
	return s.NodeGroups, nil

}

// ModifyKubernetesNodeGroup modifies the node group
func (s *UpCloudService) ModifyKubernetesNodeGroup(ctx context.Context, r *request.ModifyKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error) {
	for i := range s.NodeGroups {
		if s.NodeGroups[i].Name == r.Name {
			s.NodeGroups[i].Count = r.NodeGroup.Count
			return &s.NodeGroups[i], nil
		}
	}
	return nil, fmt.Errorf("node group not found %+v", r)
}

// DeleteKubernetesNodeGroupNode deletes the node group
func (s *UpCloudService) DeleteKubernetesNodeGroupNode(ctx context.Context, r *request.DeleteKubernetesNodeGroupNodeRequest) error {
	return nil
}

// GetKubernetesNodeGroup returns node group details
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

// GetKubernetesCluster return UKS cluster object
func (s *UpCloudService) GetKubernetesCluster(ctx context.Context, r *request.GetKubernetesClusterRequest) (*upcloud.KubernetesCluster, error) {
	for i := range s.Clusters {
		if s.Clusters[i].UUID == r.UUID {
			return &s.Clusters[i], nil
		}
	}
	return nil, &upcloud.Problem{Status: http.StatusNotFound}
}

// GetKubernetesPlans list UKS plans
func (s *UpCloudService) GetKubernetesPlans(ctx context.Context, r *request.GetKubernetesPlansRequest) ([]upcloud.KubernetesPlan, error) {
	return s.Plans, nil
}

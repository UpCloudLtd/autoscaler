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

package upcloud

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/mocks"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

func TestUpCloudCloudProvider_NodeGroups(t *testing.T) {
	t.Parallel()

	svc := mocks.UpCloudService{}
	clusterID := uuid.New()
	p := newUpCloudCloudProvider(clusterID, &svc)
	require.NoError(t, p.Refresh())
	svc.NodeGroups = append(svc.NodeGroups, upcloud.KubernetesNodeGroup{Count: 3, Name: "group3"})
	// node group length should still be 2 as refresh is not yet called
	require.Len(t, p.NodeGroups(), 2)
	require.NoError(t, p.Refresh())
	require.Len(t, p.NodeGroups(), 3)
}

func TestUpCloudCloudProvider_Name(t *testing.T) {
	t.Parallel()

	p := UpCloudCloudProvider{}
	require.Equal(t, cloudprovider.UpCloudProviderName, p.Name())
}

func TestUpCloudCloudProvider_NodeGroupForNode(t *testing.T) {
	t.Parallel()

	svc := mocks.UpCloudService{}
	clusterID := uuid.New()
	p := newUpCloudCloudProvider(clusterID, &svc)
	require.NoError(t, p.Refresh())

	group, err := p.NodeGroupForNode(&v1.Node{
		Spec: v1.NodeSpec{
			ProviderID: fmt.Sprintf("upcloud:////%s", "group1-1"),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, group)
	require.Equal(t, fmt.Sprintf("%s/group1", clusterID.String()), group.Id())

	// nodes of other providers should return `nil` as group and error
	group, err = p.NodeGroupForNode(&v1.Node{
		Spec: v1.NodeSpec{
			ProviderID: fmt.Sprintf("fake:////%s", "group1-1"),
		},
	})
	require.NoError(t, err)
	require.Nil(t, group)
}

func TestUpCloudCloudProvider_GetResourceLimiter(t *testing.T) {
	t.Parallel()

	rl := cloudprovider.NewResourceLimiter(map[string]int64{"min": 1}, nil)
	p := UpCloudCloudProvider{
		manager: &manager{
			clusterID: uuid.New(),
		},
		resourceLimiter: rl,
	}
	l, err := p.GetResourceLimiter()
	require.NoError(t, err)
	require.Equal(t, rl, l)
}

func TestBuildCloudConfig(t *testing.T) {
	want := upCloudConfig{
		ClusterID: uuid.NewString(),
		Username:  "uks-username",
		Password:  "uks-passwd",
		UserAgent: "uks-agent",
	}
	_, err := buildCloudConfig(config.AutoscalingOptions{UserAgent: want.UserAgent})
	require.Error(t, err)

	t.Setenv(envUpCloudClusterID, want.ClusterID)
	_, err = buildCloudConfig(config.AutoscalingOptions{UserAgent: want.UserAgent})
	require.Error(t, err)

	t.Setenv(envUpCloudUsername, want.Username)
	_, err = buildCloudConfig(config.AutoscalingOptions{UserAgent: want.UserAgent})
	require.Error(t, err)

	t.Setenv(envUpCloudPassword, want.Password)
	got, err := buildCloudConfig(config.AutoscalingOptions{UserAgent: want.UserAgent})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestUpCloudCloudProvider_GPULabel(t *testing.T) {
	t.Parallel()

	p := UpCloudCloudProvider{}
	require.Empty(t, p.GPULabel())
}

func TestUpCloudCloudProvider_GetAvailableGPUTypes(t *testing.T) {
	t.Parallel()

	p := UpCloudCloudProvider{}
	require.Nil(t, p.GetAvailableGPUTypes())
}

func TestUpCloudCloudProvider_Cleanup(t *testing.T) {
	t.Parallel()

	p := UpCloudCloudProvider{}
	require.Nil(t, p.Cleanup())
}

func TestUpCloudCloudProvider_GetNodeGpuConfig(t *testing.T) {
	t.Parallel()

	p := UpCloudCloudProvider{}
	require.Nil(t, p.GetNodeGpuConfig(&v1.Node{
		Spec: v1.NodeSpec{
			ProviderID: fmt.Sprintf("upcloud:////%s", uuid.NewString()),
		},
	}))
}

func TestUpCloudCloudProvider_ErrNotImplemented(t *testing.T) {
	t.Parallel()

	p := UpCloudCloudProvider{}

	_, err := p.HasInstance(nil)
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)

	_, err = p.Pricing()
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)

	_, err = p.GetAvailableMachineTypes()
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)

	_, err = p.NewNodeGroup("", nil, nil, nil, nil)
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)
}

func newUpCloudCloudProvider(clusterID uuid.UUID, svc *mocks.UpCloudService) UpCloudCloudProvider {
	if svc == nil {
		svc = &mocks.UpCloudService{}
	}
	svc.NodeGroups = append(svc.NodeGroups,
		upcloud.KubernetesNodeGroup{Count: 2, Name: "group1"},
		upcloud.KubernetesNodeGroup{Count: 3, Name: "group2"})
	return UpCloudCloudProvider{
		manager: &manager{
			clusterID: clusterID,
			svc:       svc,
		},
		resourceLimiter: cloudprovider.NewResourceLimiter(map[string]int64{"min": 1}, nil),
	}
}

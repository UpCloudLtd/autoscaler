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
)

func TestUpCloudCloudProvider(t *testing.T) {
	svc := &mocks.UpCloudService{
		NodeGroups: []upcloud.KubernetesNodeGroup{
			{Count: 2, Name: "group1"},
			{Count: 3, Name: "group2"},
		},
	}
	clusterID := uuid.New()
	p := UpCloudCloudProvider{
		manager: &manager{
			clusterID: clusterID,
			svc:       svc,
		},
		resourceLimiter: &cloudprovider.ResourceLimiter{},
	}
	require.Equal(t, cloudprovider.UpCloudProviderName, p.Name())
	require.NoError(t, p.Refresh())
	svc.NodeGroups = append(svc.NodeGroups, upcloud.KubernetesNodeGroup{Count: 3, Name: "group3"})
	// node group length should still be 2 as refresh is not yet called
	require.Len(t, p.NodeGroups(), 2)
	require.NoError(t, p.Refresh())
	require.Len(t, p.NodeGroups(), 3)

	group, err := p.NodeGroupForNode(&v1.Node{
		Spec: v1.NodeSpec{
			ProviderID: fmt.Sprintf("upcloud:////%s", "group1-1"),
		},
		Status: v1.NodeStatus{},
	})
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s/group1", clusterID.String()), group.Id())
}

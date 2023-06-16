package upcloud

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud/request"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

func TestUpCloudNodeGroup_Id(t *testing.T) {
	g := &UpCloudNodeGroup{clusterID: uuid.New(), name: "test"}
	require.Equal(t, fmt.Sprintf("%s/%s", g.clusterID.String(), g.name), g.Id())
}

func TestUpCloudNodeGroup_MinSize(t *testing.T) {
	g := &UpCloudNodeGroup{minSize: 1}
	require.Equal(t, 1, g.MinSize())
}
func TestUpCloudNodeGroup_MaxSize(t *testing.T) {
	g := &UpCloudNodeGroup{maxSize: 1}
	require.Equal(t, 1, g.MaxSize())
}
func TestUpCloudNodeGroup_TargetSize(t *testing.T) {
	g := &UpCloudNodeGroup{size: 1}
	size, err := g.TargetSize()
	require.NoError(t, err)
	require.Equal(t, 1, size)
}
func TestUpCloudNodeGroup_IncreaseSize(t *testing.T) {
	svc := upCloudServiceMock{
		nodeGroups: []upcloud.KubernetesNodeGroup{
			{
				Name:  "test",
				Count: 1,
				State: upcloud.KubernetesNodeGroupStateRunning,
			},
		},
	}
	g := &UpCloudNodeGroup{size: 1, maxSize: 20, name: "test", svc: &svc}
	require.NoError(t, g.IncreaseSize(1))
	size, _ := g.TargetSize()
	require.Equal(t, 2, size)
}
func TestUpCloudNodeGroup_DecreaseTargetSize(t *testing.T) {
	svc := upCloudServiceMock{
		nodeGroups: []upcloud.KubernetesNodeGroup{
			{
				Name:  "test",
				Count: 3,
				State: upcloud.KubernetesNodeGroupStateRunning,
			},
		},
	}
	g := &UpCloudNodeGroup{size: 3, maxSize: 20, name: "test", svc: &svc}
	require.NoError(t, g.DecreaseTargetSize(-1))
	size, _ := g.TargetSize()
	require.Equal(t, 2, size)
}
func TestUpCloudNodeGroup_DeleteNodes(t *testing.T) {
	svc := upCloudServiceMock{
		nodeGroups: []upcloud.KubernetesNodeGroup{
			{
				Name:  "test",
				Count: 1,
				State: upcloud.KubernetesNodeGroupStateRunning,
			},
		},
	}
	g := &UpCloudNodeGroup{size: 3, maxSize: 20, name: "test", svc: &svc}
	require.NoError(t, g.DeleteNodes([]*v1.Node{
		{ObjectMeta: metav1.ObjectMeta{Name: "test-1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "test-2"}},
	}))
	size, _ := g.TargetSize()
	require.Equal(t, 1, size)
}
func TestUpCloudNodeGroup_Nodes(t *testing.T) {
	wantNodes := []cloudprovider.Instance{{
		Id: "test",
	}}
	g := &UpCloudNodeGroup{nodes: wantNodes}
	gotNodes, err := g.Nodes()
	require.NoError(t, err)
	require.Equal(t, wantNodes, gotNodes)
}
func TestUpCloudNodeGroup_Autoprovisioned(t *testing.T) {
	g := &UpCloudNodeGroup{}
	require.False(t, g.Autoprovisioned())
}
func TestUpCloudNodeGroup_Create(t *testing.T) {
	g := &UpCloudNodeGroup{}
	_, err := g.Create()
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)
}
func TestUpCloudNodeGroup_Delete(t *testing.T) {
	g := &UpCloudNodeGroup{}
	err := g.Delete()
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)
}
func TestUpCloudNodeGroup_GetOptions(t *testing.T) {
	g := &UpCloudNodeGroup{}
	_, err := g.GetOptions(config.NodeGroupAutoscalingOptions{})
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)
}
func TestUpCloudNodeGroup_Debug(t *testing.T) {
	g := &UpCloudNodeGroup{name: "test"}
	require.NotEmpty(t, g.Debug())
}
func TestUpCloudNodeGroup_Exist(t *testing.T) {
	g := &UpCloudNodeGroup{name: "test"}
	require.True(t, g.Exist())
}
func TestUpCloudNodeGroup_TemplateNodeInfo(t *testing.T) {
	g := &UpCloudNodeGroup{}
	_, err := g.TemplateNodeInfo()
	require.ErrorIs(t, err, cloudprovider.ErrNotImplemented)
}

type upCloudServiceMock struct {
	nodeGroups         []upcloud.KubernetesNodeGroup
	detailedNodeGroups []upcloud.KubernetesNodeGroupDetails
}

func (s *upCloudServiceMock) GetKubernetesNodeGroups(ctx context.Context, r *request.GetKubernetesNodeGroupsRequest) ([]upcloud.KubernetesNodeGroup, error) {
	return s.nodeGroups, nil

}
func (s *upCloudServiceMock) ModifyKubernetesNodeGroup(ctx context.Context, r *request.ModifyKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error) {
	for i := range s.nodeGroups {
		if s.nodeGroups[i].Name == r.Name {
			s.nodeGroups[i].Count = r.NodeGroup.Count
			return &s.nodeGroups[i], nil
		}
	}
	return nil, fmt.Errorf("node group not found %+v", r)
}

func (s *upCloudServiceMock) DeleteKubernetesNodeGroupNode(ctx context.Context, r *request.DeleteKubernetesNodeGroupNodeRequest) error {
	return nil
}

func (s *upCloudServiceMock) GetKubernetesNodeGroup(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroupDetails, error) {
	for i := range s.detailedNodeGroups {
		g := s.detailedNodeGroups[i]
		if g.Name == r.Name {
			return &g, nil
		}
	}
	return nil, fmt.Errorf("node group details not found %+v", r)
}

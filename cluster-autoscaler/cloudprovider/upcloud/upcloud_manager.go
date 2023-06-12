package upcloud

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud/client"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud/request"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud/service"
	"k8s.io/klog/v2"
)

type upCloudService interface {
	GetKubernetesNodeGroups(ctx context.Context, r *request.GetKubernetesNodeGroupsRequest) ([]upcloud.KubernetesNodeGroup, error)
	GetKubernetesNodeGroup(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error)
	ModifyKubernetesNodeGroup(ctx context.Context, r *request.ModifyKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error)
	DeleteKubernetesNodeGroupNode(ctx context.Context, r *request.DeleteKubernetesNodeGroupNodeRequest) error
	GetKubernetesNodeGroupDetails(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroupDetails, error)
}

type Manager struct {
	clusterID uuid.UUID

	svc        upCloudService
	nodeGroups []UpCloudNodeGroup
	mu         sync.Mutex
}

func (m *Manager) Refresh() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), timeoutGetRequest)
	defer cancel()
	groups := make([]UpCloudNodeGroup, 0)
	upcloudNodeGroups, err := m.svc.GetKubernetesNodeGroups(ctx, &request.GetKubernetesNodeGroupsRequest{
		ClusterUUID: m.clusterID.String(),
	})
	if err != nil {
		return err
	}
	for _, g := range upcloudNodeGroups {
		nodes, err := nodeGroupNodes(m.svc, m.clusterID, g.Name)
		if err != nil {
			klog.ErrorS(err, "failed to get node group nodes")
			continue
		}
		group := UpCloudNodeGroup{
			clusterID: m.clusterID,
			name:      g.Name,
			size:      g.Count,
			minSize:   nodeGroupMinSize,
			maxSize:   nodeGroupMaxSize,
			svc:       m.svc,
			nodes:     nodes,
		}
		klog.V(logInfo).Infof("caching cluster %s node group %s size=%d minSize=%d maxSize=%d nodes=%d",
			m.clusterID.String(), group.name, group.size, group.minSize, group.maxSize, len(groups))
		groups = append(groups, group)
	}
	m.nodeGroups = groups
	klog.V(logInfo).Infof("refreshed node groups (%d)", len(m.nodeGroups))
	return nil
}

func newManager() (*Manager, error) {
	const (
		envUpCloudUsername  string = "UPCLOUD_USERNAME"
		envUpCloudPassword  string = "UPCLOUD_PASSWORD"
		envUpCloudClusterID string = "UPCLOUD_CLUSTER_ID"
	)
	var (
		upCloudUsername, upCloudPassword, upCloudClusterID string
	)
	if upCloudUsername = os.Getenv(envUpCloudUsername); upCloudUsername == "" {
		return nil, fmt.Errorf("environment variable %s not set", envUpCloudUsername)
	}
	if upCloudPassword = os.Getenv(envUpCloudPassword); upCloudPassword == "" {
		return nil, fmt.Errorf("environment variable %s not set", envUpCloudPassword)
	}
	if upCloudClusterID = os.Getenv(envUpCloudClusterID); upCloudClusterID == "" {
		return nil, fmt.Errorf("environment variable %s not set", envUpCloudClusterID)
	}
	clusterID, err := uuid.Parse(upCloudClusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster ID %s is not valid UUID %w", envUpCloudClusterID, err)
	}
	return &Manager{
		clusterID:  clusterID,
		svc:        service.New(client.New(upCloudUsername, upCloudPassword)),
		nodeGroups: make([]UpCloudNodeGroup, 0),
		mu:         sync.Mutex{},
	}, nil
}

func nodeGroupNodes(svc upCloudService, clusterID uuid.UUID, name string) ([]cloudprovider.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutGetRequest)
	defer cancel()
	instances := make([]cloudprovider.Instance, 0)
	klog.V(logInfo).Infof("fetching node group %s/%s details", clusterID.String(), name)
	ng, err := svc.GetKubernetesNodeGroupDetails(ctx, &request.GetKubernetesNodeGroupRequest{
		ClusterUUID: clusterID.String(),
		Name:        name,
	})
	if err != nil {
		return instances, err
	}
	for i := range ng.Nodes {
		node := ng.Nodes[i]
		instances = append(instances, cloudprovider.Instance{
			Id:     fmt.Sprintf("upcloud:////%s", node.UUID),
			Status: nodeStateToInstanceStatus(node.State),
		})
	}
	return instances, err
}

func nodeStateToInstanceStatus(nodeState upcloud.KubernetesNodeState) *cloudprovider.InstanceStatus {
	var s cloudprovider.InstanceState
	var e *cloudprovider.InstanceErrorInfo
	switch nodeState {
	case upcloud.KubernetesNodeStateRunning:
		s = cloudprovider.InstanceRunning
	case upcloud.KubernetesNodeStateTerminating:
		s = cloudprovider.InstanceDeleting
	case upcloud.KubernetesNodeStatePending:
		s = cloudprovider.InstanceCreating
	default:
		e = &cloudprovider.InstanceErrorInfo{
			ErrorClass: cloudprovider.OtherErrorClass,
			ErrorCode:  string(nodeState),
		}
	}
	return &cloudprovider.InstanceStatus{
		State:     s,
		ErrorInfo: e,
	}
}

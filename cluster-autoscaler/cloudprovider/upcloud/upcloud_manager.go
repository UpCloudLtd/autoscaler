package upcloud

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/v6/upcloud/service"
	"github.com/google/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog/v2"
)

type upCloudService interface {
	GetKubernetesNodeGroups(ctx context.Context, r *request.GetKubernetesNodeGroupsRequest) ([]upcloud.KubernetesNodeGroup, error)
	GetKubernetesNodeGroup(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error)
	ModifyKubernetesNodeGroup(ctx context.Context, r *request.ModifyKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error)
	GetServerGroups(ctx context.Context, r *request.GetServerGroupsRequest) (upcloud.ServerGroups, error)
	GetServerDetails(ctx context.Context, r *request.GetServerDetailsRequest) (*upcloud.ServerDetails, error)
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
		klog.V(4).Infof("caching cluster %s node group %s size=%d minSize=%d maxSize=%d nodes=%d",
			m.clusterID.String(), group.name, group.size, group.minSize, group.maxSize, len(groups))
		groups = append(groups, group)
	}
	m.nodeGroups = groups
	klog.V(4).Infof("refreshed node groups (%d)", len(m.nodeGroups))
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
	n := make([]cloudprovider.Instance, 0)

	klog.V(4).Infof("fetching server group to determine %s/%s nodes", clusterID.String(), name)
	serverGroups, err := svc.GetServerGroups(ctx, &request.GetServerGroupsRequest{
		Filters: []request.QueryFilter{
			request.FilterLabel{
				Label: upcloud.Label{
					Key:   "capu_cluster_id",
					Value: clusterID.String(),
				},
			},
			request.FilterLabel{
				Label: upcloud.Label{
					Key:   "capu_generated_name",
					Value: fmt.Sprintf("%s-server-group", name),
				},
			},
		},
	})
	if err != nil {
		return n, err
	}
	if len(serverGroups) != 1 {
		return n, fmt.Errorf("got ambiguous server groups response, wanted 1 got %d", len(serverGroups))
	}
	serverGroup := serverGroups[0]
	klog.V(4).Infof("%s/%s found server group %s with %d members [%s]",
		clusterID.String(), name, serverGroup.Title, len(serverGroup.Members), strings.Join(serverGroup.Members, ","))

	for _, serverID := range serverGroup.Members {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutGetRequest)
		defer cancel()
		srv, err := svc.GetServerDetails(ctx, &request.GetServerDetailsRequest{UUID: serverID})
		if err != nil {
			return n, fmt.Errorf("failed to get node details, %w", err)
		}
		n = append(n, cloudprovider.Instance{
			Id:     fmt.Sprintf("upcloud:////%s", serverID),
			Status: serverStateToInstanceStatus(srv.State),
		})
	}

	return n, nil
}

func serverStateToInstanceStatus(serverStatus string) *cloudprovider.InstanceStatus {
	var s cloudprovider.InstanceState
	var e *cloudprovider.InstanceErrorInfo
	switch serverStatus {
	case "started":
		s = cloudprovider.InstanceRunning
	case "maintenance", "stopped":
		s = cloudprovider.InstanceDeleting
	default:
		// error
		e = &cloudprovider.InstanceErrorInfo{
			ErrorClass: cloudprovider.OtherErrorClass,
			ErrorCode:  serverStatus,
		}
	}
	return &cloudprovider.InstanceStatus{
		State:     s,
		ErrorInfo: e,
	}
}

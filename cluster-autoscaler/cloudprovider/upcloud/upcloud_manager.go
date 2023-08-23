package upcloud

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud/request"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/klog/v2"
)

type upCloudService interface {
	GetKubernetesCluster(ctx context.Context, r *request.GetKubernetesClusterRequest) (*upcloud.KubernetesCluster, error)
	GetKubernetesNodeGroups(ctx context.Context, r *request.GetKubernetesNodeGroupsRequest) ([]upcloud.KubernetesNodeGroup, error)
	GetKubernetesNodeGroup(ctx context.Context, r *request.GetKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroupDetails, error)
	ModifyKubernetesNodeGroup(ctx context.Context, r *request.ModifyKubernetesNodeGroupRequest) (*upcloud.KubernetesNodeGroup, error)
	DeleteKubernetesNodeGroupNode(ctx context.Context, r *request.DeleteKubernetesNodeGroupNodeRequest) error
	GetKubernetesPlans(ctx context.Context, r *request.GetKubernetesPlansRequest) ([]upcloud.KubernetesPlan, error)
}

// manager manages node group cache
type manager struct {
	clusterID  uuid.UUID
	svc        upCloudService
	nodeGroups []*UpCloudNodeGroup

	maxNodesTotal int

	mu sync.Mutex
}

// refresh updates manager's node group cache
func (m *manager) refresh() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), timeoutGetRequest)
	defer cancel()
	groups := make([]*UpCloudNodeGroup, 0)
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
			maxSize:   m.maxNodesTotal,
			svc:       m.svc,
			nodes:     nodes,
			mu:        sync.Mutex{},
		}
		klog.V(logInfo).Infof("caching cluster %s node group %s size=%d minSize=%d maxSize=%d nodes=%d",
			m.clusterID.String(), group.name, group.size, group.minSize, group.maxSize, len(nodes))
		groups = append(groups, &group)
	}
	m.nodeGroups = groups
	klog.V(logInfo).Infof("refreshed node groups (%d)", len(m.nodeGroups))
	return nil
}

func newManager(ctx context.Context, svc upCloudService, cfg upCloudConfig, opts config.AutoscalingOptions) (*manager, error) {
	clusterUUID, err := uuid.Parse(cfg.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("cluster ID %s is not valid UUID %w", envUpCloudClusterID, err)
	}

	maxNodesTotal, err := clusterMaxNodes(ctx, svc, clusterUUID, opts.MaxNodesTotal)
	if err != nil {
		return nil, err
	}
	return &manager{
		clusterID:     clusterUUID,
		maxNodesTotal: maxNodesTotal,
		svc:           svc,
		nodeGroups:    make([]*UpCloudNodeGroup, 0),
		mu:            sync.Mutex{},
	}, nil
}

func clusterMaxNodes(ctx context.Context, svc upCloudService, clusterID uuid.UUID, requestedMaxNodesTotal int) (int, error) {
	cluster, err := svc.GetKubernetesCluster(ctx, &request.GetKubernetesClusterRequest{
		UUID: clusterID.String(),
	})
	if err != nil {
		var p *upcloud.Problem
		if err != nil && errors.As(err, &p) && p.Status == http.StatusForbidden {
			return requestedMaxNodesTotal, fmt.Errorf("unable to get cluster %s info, permission defined", clusterID.String())
		}
		return requestedMaxNodesTotal, err
	}

	plan, err := clusterPlanByName(ctx, svc, cluster.Plan)
	if err != nil {
		return requestedMaxNodesTotal, err
	}

	if requestedMaxNodesTotal == 0 {
		return plan.MaxNodes, nil
	}
	if requestedMaxNodesTotal > plan.MaxNodes {
		return requestedMaxNodesTotal, fmt.Errorf("MaxNodesTotal %d is greater than maximum allowed %d with selected %s cluster plan", requestedMaxNodesTotal, plan.MaxNodes, cluster.Plan)
	}
	return requestedMaxNodesTotal, nil
}

func clusterPlanByName(ctx context.Context, svc upCloudService, name string) (upcloud.KubernetesPlan, error) {
	plans, err := svc.GetKubernetesPlans(ctx, &request.GetKubernetesPlansRequest{})
	if err != nil {
		return upcloud.KubernetesPlan{}, err
	}
	for i := range plans {
		if strings.EqualFold(plans[i].Name, name) {
			return plans[i], nil
		}
	}
	return upcloud.KubernetesPlan{}, fmt.Errorf("can't get cluster plan by name '%s'", name)
}

func nodeGroupNodes(svc upCloudService, clusterID uuid.UUID, name string) ([]cloudprovider.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutGetRequest)
	defer cancel()
	instances := make([]cloudprovider.Instance, 0)
	klog.V(logInfo).Infof("fetching node group %s/%s details", clusterID.String(), name)
	ng, err := svc.GetKubernetesNodeGroup(ctx, &request.GetKubernetesNodeGroupRequest{
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

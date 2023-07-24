package upcloud

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud/request"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type UpCloudNodeGroup struct {
	// implements cloudprovide.NodeGroup interfaces
	clusterID uuid.UUID
	name      string
	size      int
	minSize   int
	maxSize   int

	nodes []cloudprovider.Instance
	svc   upCloudService

	mu sync.Mutex
}

// Id returns an unique identifier of the node group.
func (u *UpCloudNodeGroup) Id() string {
	id := fmt.Sprintf("%s/%s", u.clusterID.String(), u.name)
	// set log level higher because this get called a lot
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Id called", id)
	return id
}

// MinSize returns minimum size of the node group.
func (u *UpCloudNodeGroup) MinSize() int {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.MinSize called", u.Id())
	return u.minSize
}

// MaxSize returns maximum size of the node group.
func (u *UpCloudNodeGroup) MaxSize() int {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.MaxSize called", u.Id())
	return u.maxSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (u *UpCloudNodeGroup) TargetSize() (int, error) {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.TargetSize called (%d)", u.Id(), u.size)
	return u.size, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (u *UpCloudNodeGroup) IncreaseSize(delta int) error {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.IncreaseSize(%d) called", u.Id(), delta)
	if delta <= 0 {
		return fmt.Errorf("failed to increase node group size, delta=%d", delta)
	}
	size := u.size + delta
	if size > u.MaxSize() {
		return fmt.Errorf("failed to increase node group size, current=%d want=%d max=%d", u.size, size, u.MaxSize())
	}
	return u.scaleNodeGroup(size)
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (u *UpCloudNodeGroup) DecreaseTargetSize(delta int) error {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.DecreaseTargetSize(%d) called", u.Id(), delta)
	if delta >= 0 {
		return fmt.Errorf("failed to increase node group size, delta=%d", delta)
	}
	size := u.size + delta
	if size < u.MinSize() {
		return fmt.Errorf("failed to decrease node group size, current=%d want=%d min=%d", u.size, size, u.MinSize())
	}
	return u.scaleNodeGroup(size)
}

func (u *UpCloudNodeGroup) scaleNodeGroup(size int) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), timeoutModifyNodeGroup)
	defer cancel()
	klog.V(logInfo).Infof("scaling node group %s from %d to %d", u.Id(), u.size, size)
	_, err := u.svc.ModifyKubernetesNodeGroup(ctx, &request.ModifyKubernetesNodeGroupRequest{
		ClusterUUID: u.clusterID.String(),
		Name:        u.name,
		NodeGroup: request.ModifyKubernetesNodeGroup{
			Count: size,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to scale node group %s, %w", u.name, err)
	}
	nodeGroup, err := u.waitNodeGroupState(upcloud.KubernetesNodeGroupStateRunning, timeoutWaitNodeGroupState)
	if err != nil {
		return err
	}
	u.size = nodeGroup.Count
	return nil
}

func (u *UpCloudNodeGroup) waitNodeGroupState(state upcloud.KubernetesNodeGroupState, timeout time.Duration) (*upcloud.KubernetesNodeGroupDetails, error) {
	deadline := time.Now().Add(timeout)
	i := 1
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutGetRequest)
		defer cancel()
		klog.V(logInfo).Infof("waiting(%d) node group %s state %s", i, u.Id(), state)
		g, err := u.svc.GetKubernetesNodeGroup(ctx, &request.GetKubernetesNodeGroupRequest{
			ClusterUUID: u.clusterID.String(),
			Name:        u.name,
		})
		if err != nil {
			return g, fmt.Errorf("failed to fetch node group %s, %w", u.Id(), err)
		}
		if g.State == state {
			return g, nil
		}
		time.Sleep(2 * time.Second)
		i++
	}
	return nil, fmt.Errorf("node group %s state check (%d) timed out", u.Id(), i)
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (u *UpCloudNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.DeleteNodes called", u.Id())
	u.mu.Lock()
	defer u.mu.Unlock()
	for i := range nodes {
		if err := u.deleteNode(nodes[i].GetName()); err != nil {
			return err
		}
	}
	nodeGroup, err := u.waitNodeGroupState(upcloud.KubernetesNodeGroupStateRunning, timeoutWaitNodeGroupState)
	if err != nil {
		return err
	}
	u.size = nodeGroup.Count
	return nil
}

func (u *UpCloudNodeGroup) deleteNode(nodeName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutDeleteNode)
	defer cancel()
	klog.V(logInfo).Infof("deleting UpCloud %s/node %s", u.Id(), nodeName)
	return u.svc.DeleteKubernetesNodeGroupNode(ctx, &request.DeleteKubernetesNodeGroupNodeRequest{
		ClusterUUID: u.clusterID.String(),
		Name:        u.name,
		NodeName:    nodeName,
	})
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (u *UpCloudNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Nodes called", u.Id())
	return u.nodes, nil
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (u *UpCloudNodeGroup) Autoprovisioned() bool {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Autoprovisioned called", u.Id())
	return false
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (u *UpCloudNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Create called", u.Id())
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (u *UpCloudNodeGroup) Delete() error {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Delete called", u.Id())
	return cloudprovider.ErrNotImplemented
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
// Implementation optional.
func (u *UpCloudNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.GetOptions called", u.Id())
	return nil, cloudprovider.ErrNotImplemented
}

// Debug returns a string containing all information regarding this node group.
func (u *UpCloudNodeGroup) Debug() string {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Debug called", u.Id())
	return fmt.Sprintf("Node group ID: %s (min:%d max:%d)", u.Id(), u.MinSize(), u.MaxSize())
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (u *UpCloudNodeGroup) Exist() bool {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.Exist called", u.Id())
	return u.name != ""
}

// TemplateNodeInfo returns a schedulerframework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (u *UpCloudNodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	klog.V(logDebug).Infof("UpCloud %s/NodeGroup.TemplateNodeInfo called", u.Id())
	return nil, cloudprovider.ErrNotImplemented
}

package request

import "fmt"

type DeleteKubernetesNodeGroupNodeRequest struct {
	ClusterUUID   string
	NodeGroupName string
	Name          string
}

func (r *DeleteKubernetesNodeGroupNodeRequest) RequestURL() string {
	return fmt.Sprintf("%s/%s/node-groups/%s/%s", kubernetesClusterBasePath, r.ClusterUUID, r.NodeGroupName, r.Name)
}

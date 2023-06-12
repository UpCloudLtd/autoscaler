package upcloud

type KubernetesNodeState string

const (
	KubernetesNodeStateNodeStateFailed KubernetesNodeState = "failed"
	KubernetesNodeStatePending         KubernetesNodeState = "pending"
	KubernetesNodeStateRunning         KubernetesNodeState = "running"
	KubernetesNodeStateTerminating     KubernetesNodeState = "terminating"
	KubernetesNodeStateUnknown         KubernetesNodeState = "unknown"
)

type KubernetesNode struct {
	UUID  string              `json:"uuid,omitempty"`
	Name  string              `json:"name,omitempty"`
	State KubernetesNodeState `json:"state,omitempty"`
}

type KubernetesNodeGroupDetails struct {
	AntiAffinity bool                     `json:"anti_affinity,omitempty"`
	Count        int                      `json:"count,omitempty"`
	KubeletArgs  []KubernetesKubeletArg   `json:"kubelet_args,omitempty"`
	Labels       []Label                  `json:"labels,omitempty"`
	Name         string                   `json:"name,omitempty"`
	Plan         string                   `json:"plan,omitempty"`
	SSHKeys      []string                 `json:"ssh_keys,omitempty"`
	State        KubernetesNodeGroupState `json:"state,omitempty"`
	Storage      string                   `json:"storage,omitempty"`
	Taints       []KubernetesTaint        `json:"taints,omitempty"`
	Nodes        []KubernetesNode         `json:"nodes,omitempty"`
}

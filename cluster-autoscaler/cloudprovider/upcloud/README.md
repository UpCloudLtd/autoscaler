# Cluster Autoscaler for UpCloud (beta)

# Overview

Cluster Autoscaler for UpCloud automatically adjusts the size of UKS node groups when one of the following conditions is true: 
- there are pods that failed to run in the cluster due to insufficient resources.
- there are nodes in the cluster that have been underutilized for an extended period of time and their pods can be placed on other existing nodes.

Additional info about the Cluster Autoscaler (CA) can be found from the project's [README](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/README.md) file and from the [FAQ](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md) .

Latest Docker image is available at Github package registry
```shell
$ docker pull ghcr.io/upcloudltd/autoscaler:latest
```

## Configuration
### Required environment variables
- `UPCLOUD_USERNAME` - UpCloud's API username
- `UPCLOUD_PASSWORD` - UpCloud's API user's password
- `UPCLOUD_CLUSTER_ID` - UKS cluster ID

### Optional environment variables
- `UPCLOUD_DEBUG_API_BASE_URL` - Use alternative UpCloud API URL

## Build
Go to `autoscaler/cluster-autoscaler` directory  

build binary:
```shell
$ BUILD_TAGS=upcloud make build-in-docker
```

build image:
```shell
$ docker build -t <image:tag> -f Dockerfile.amd64 .
```

## Deployment

### Create a Kubernetes secret
Execute the following command to add UpCloud credentials as Kubernetes secret:  
<sub>_Replace `$UPCLOUD_PASSWORD` and `$UPCLOUD_USERNAME` with your UpCloud API credentials if not defined using environment variables._</sub>
```shell
$ kubectl -n kube-system create secret generic upcloud-autoscaler --from-literal=password=$UPCLOUD_PASSWORD --from-literal=username=$UPCLOUD_USERNAME
```
Note that user `$UPCLOUD_USERNAME` needs to have permission to manage Kubernetes cluster through UpCloud API.

### Deploy Cluster Autoscaler
Update your UKS cluster ID (`UPCLOUD_CLUSTER_ID`) into [examples/cluster-autoscaler.yaml](./examples/cluster-autoscaler.yaml)

```shell
$ kubectl apply -f examples/rbac.yaml
$ kubectl apply -f examples/cluster-autoscaler.yaml
```

### Customize node group limits
By default each node group have at least one node and at most the number of nodes that are allowed in selected cluster plan.
These limits can be customized with `--nodes` command-line argument, using format `<min>:<max>:<node_group_name>`.

For example to make sure that:
- node group `monitor` has always at least 2 nodes, but never over 10 nodes
- node group `dev` has always at least 2 nodes, but never over 3 nodes
- rest of the node groups can scale up and down freely

*Note that CA will not scale __up__ node group automatically to minimum value, minimum value only effects scale __down__ operation.*

```yaml
command:
    - /cluster-autoscaler
    - --cloud-provider=upcloud
    - --stderrthreshold=info
    - --scale-down-enabled=true
    - --v=4
    - --nodes=2:10:monitor
    - --nodes=2:3:dev
```


## Test scaling up

Deploy example app
```shell
$ kubectl apply -f examples/testing/deployment.yaml
```
Increase app replicas (e.g. 20-50) until you see node group scaling up.

## Run locally using kubeconfig file 
Build `autoscaler/cluster-autoscaler/cluster-autoscaler-amd64` binary
```shell
$ make build
```

Setup environment variables and run autoscaler binary:
```shell
$ ./cluster-autoscaler-amd64 --address=:8087 --cloud-provider=upcloud --stderrthreshold=info --scale-down-enabled=true --v=4 --kubeconfig=<path to kubeconfig file>
```
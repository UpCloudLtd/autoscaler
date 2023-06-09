# Cluster Autoscaler for UpCloud (experimental)

This is just experimental implementation and it's not working as intended yet.

## Todo
- [x] vendor UpCloud Go SDK
- [ ] implement NodeGroup.DeleteNodes() - scaling down is not working
- [ ] clean up code and fix bugs
- [ ] add tests

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
Update your UKS cluster ID (`UPCLOUD_CLUSTER_ID`) into [examples/cluster-autoscaler.yaml](./examples/cluster-autoscaler.yaml)

```shell
$ kubectl apply -f examples/rbac.yaml
$ kubectl apply -f examples/cluster-autoscaler.yaml
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
$ ./cluster-autoscaler-amd64 --address=:8087 --cloud-provider=upcloud --stderrthreshold=info --scale-down-enabled=false --v=4 --kubeconfig=<path to kubeconfig file>
```
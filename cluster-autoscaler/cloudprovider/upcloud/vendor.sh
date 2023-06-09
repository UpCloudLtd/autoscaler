#!/bin/sh

UPCLOUD_SDK_PACKAGE=$1
UPCLOUD_SDK_VERSION=$2

if [ -z "${UPCLOUD_SDK_PACKAGE}" ] || [ -z "${UPCLOUD_SDK_VERSION}" ]; then
    echo "Usage: $0 <package> <package version>"
    exit 1
fi

sdk_dir=`dirname $0`/pkg/${UPCLOUD_SDK_PACKAGE}/upcloud
sdk_url=https://raw.githubusercontent.com/UpCloudLtd/upcloud-go-api/${UPCLOUD_SDK_VERSION}

mkdir -p $sdk_dir/client $sdk_dir/service $sdk_dir/request

sdk_download () {
    echo "${2} => ${1}"
    curl -sO --output-dir $1 $2
}

sdk_download $sdk_dir "${sdk_url}/upcloud/{kubernetes.go,network.go,problem.go,label.go,utils.go,ip_address.go}"
sdk_download $sdk_dir/client "${sdk_url}/upcloud/client/{client,error}.go"
sdk_download $sdk_dir/request "${sdk_url}/upcloud/request/{kubernetes.go,network.go,request.go,label.go}"
sdk_download $sdk_dir/service "${sdk_url}/upcloud/service/{kubernetes.go,service.go,network.go}"


# TODO: remove when UKS node API endpoints are available and append ServerGroup and Server interfaces to stubs.go
sdk_download $sdk_dir "${sdk_url}/upcloud/{server_group.go,server.go,storage.go}"
sdk_download $sdk_dir/request "${sdk_url}/upcloud/request/{server_group.go,server.go,storage.go}"
sdk_download $sdk_dir/service "${sdk_url}/upcloud/service/{server_group.go,server.go}"


echo "
package service

type Cloud interface {}
type Account interface {}
type Firewall interface {}
type Host interface {}
type IPAddress interface {}
type LoadBalancer interface {}
type Tag interface {} 
type Storage interface {} 
type ObjectStorage interface {}
type ManagedDatabaseServiceManager interface {}
type ManagedDatabaseUserManager interface {}
type ManagedDatabaseLogicalDatabaseManager interface {}
type Permission interface {}
" > $sdk_dir/service/stubs.go

#github.com/UpCloudLtd/upcloud-go-api/v6 -> "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/UpCloudLtd/upcloud-go-api/v6"
# "${UPCLOUD_SDK_PACKAGE} -> "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/${UPCLOUD_SDK_PACKAGE}
find $sdk_dir -name "*.go" -exec sed -i 's#"'${UPCLOUD_SDK_PACKAGE}'#"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/'${UPCLOUD_SDK_PACKAGE}'#gI' {} \;

#!/bin/sh
# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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

sdk_download $sdk_dir "${sdk_url}/upcloud/{kubernetes.go,problem.go,utils.go,label.go,ip_address.go,network.go,plan.go,price.go,storage.go,timezone.go,zone.go}"
sdk_download $sdk_dir/client "${sdk_url}/upcloud/client/{client,error}.go"
sdk_download $sdk_dir/request "${sdk_url}/upcloud/request/{kubernetes.go,request.go,network.go}"
sdk_download $sdk_dir/service "${sdk_url}/upcloud/service/{kubernetes.go,service.go,network.go,cloud.go,retry.go}"

echo "
package service

type Account interface{}
type Firewall interface{}
type Host interface{}
type IPAddress interface{}
type LoadBalancer interface{}
type Tag interface{}
type Storage interface{}
type ObjectStorage interface{}
type ManagedDatabaseServiceManager interface{}
type ManagedDatabaseUserManager interface{}
type ManagedDatabaseLogicalDatabaseManager interface{}
type Permission interface{}
type ServerGroup interface{}
type Server interface{}
type ManagedObjectStorage interface{}
type Gateway interface{}
type Partner interface{}
" > $sdk_dir/service/stubs.go

find $sdk_dir -name "*.go" -exec sed -i 's#"'${UPCLOUD_SDK_PACKAGE}'#"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/'${UPCLOUD_SDK_PACKAGE}'#gI' {} \;

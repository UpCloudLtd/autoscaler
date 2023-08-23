/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package upcloud

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/mocks"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud"
)

func TestClusterMaxNodes(t *testing.T) {
	t.Parallel()

	clusterUUID := uuid.New()
	mock := mocks.UpCloudService{
		Clusters: []upcloud.KubernetesCluster{{
			UUID: clusterUUID.String(),
			Plan: "dev",
		}},
		Plans: []upcloud.KubernetesPlan{{
			Name:     "dev",
			MaxNodes: 20,
		}},
	}
	want := 10
	got, err := clusterMaxNodes(context.TODO(), &mock, clusterUUID, 10)
	require.NoError(t, err)
	require.Equal(t, want, got)

	got, err = clusterMaxNodes(context.TODO(), &mock, clusterUUID, 0)
	require.NoError(t, err)
	require.Equal(t, mock.Plans[0].MaxNodes, got)

	_, err = clusterMaxNodes(context.TODO(), &mock, clusterUUID, 100)
	require.Error(t, err)
}

func TestClusterPlanByName(t *testing.T) {
	t.Parallel()

	want := upcloud.KubernetesPlan{
		Name: "match",
	}
	mock := mocks.UpCloudService{
		Plans: []upcloud.KubernetesPlan{want},
	}
	got, err := clusterPlanByName(context.TODO(), &mock, want.Name)
	require.NoError(t, err)
	require.Equal(t, want.Name, got.Name)
}

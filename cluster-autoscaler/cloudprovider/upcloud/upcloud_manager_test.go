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

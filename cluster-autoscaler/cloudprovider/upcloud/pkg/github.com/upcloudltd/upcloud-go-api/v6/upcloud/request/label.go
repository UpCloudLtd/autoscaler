package request

import (
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v6/upcloud"
)

type FilterLabel struct {
	upcloud.Label
}

func (l FilterLabel) ToQueryParam() string {
	return fmt.Sprintf("label=%s=%s", l.Key, l.Value)
}

type FilterLabelKey struct {
	Key string
}

func (k FilterLabelKey) ToQueryParam() string {
	return fmt.Sprintf("label=%s", k.Key)
}

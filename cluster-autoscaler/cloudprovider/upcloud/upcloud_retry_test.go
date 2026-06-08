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
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/upcloud/pkg/github.com/upcloudltd/upcloud-go-api/v8/upcloud/client"
)

func TestRetryableError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"deadline exceeded", context.DeadlineExceeded, true},
		{"wrapped deadline", fmt.Errorf("get node-groups: %w", context.DeadlineExceeded), true},
		{"canceled", context.Canceled, false},
		{"problem 429", &upcloud.Problem{Status: http.StatusTooManyRequests}, true},
		{"problem 500", &upcloud.Problem{Status: http.StatusInternalServerError}, true},
		{"problem 502", &upcloud.Problem{Status: http.StatusBadGateway}, true},
		{"problem 400", &upcloud.Problem{Status: http.StatusBadRequest}, false},
		{"problem 403", &upcloud.Problem{Status: http.StatusForbidden}, false},
		{"problem 404", &upcloud.Problem{Status: http.StatusNotFound}, false},
		{"client 503", &client.Error{ErrorCode: http.StatusServiceUnavailable}, true},
		{"client 429", &client.Error{ErrorCode: http.StatusTooManyRequests}, true},
		{"client 401", &client.Error{ErrorCode: http.StatusUnauthorized}, false},
		{"unknown transport error", errors.New("connection refused"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, retryableError(tt.err))
		})
	}
}

func TestWithRetry(t *testing.T) {
	old := retryBaseBackoff
	retryBaseBackoff = time.Millisecond
	t.Cleanup(func() { retryBaseBackoff = old })

	t.Run("succeeds on first attempt", func(t *testing.T) {
		calls := 0
		err := withRetry(context.Background(), time.Second, func(context.Context) error {
			calls++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 1, calls)
	})

	t.Run("retries transient errors then succeeds", func(t *testing.T) {
		calls := 0
		err := withRetry(context.Background(), time.Second, func(context.Context) error {
			calls++
			if calls < 3 {
				return context.DeadlineExceeded
			}
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 3, calls)
	})

	t.Run("does not retry non-retryable errors", func(t *testing.T) {
		calls := 0
		err := withRetry(context.Background(), time.Second, func(context.Context) error {
			calls++
			return &upcloud.Problem{Status: http.StatusBadRequest}
		})
		require.Error(t, err)
		require.Equal(t, 1, calls)
	})

	t.Run("exhausts attempts on persistent failure", func(t *testing.T) {
		calls := 0
		err := withRetry(context.Background(), time.Second, func(context.Context) error {
			calls++
			return context.DeadlineExceeded
		})
		require.ErrorIs(t, err, context.DeadlineExceeded)
		require.Equal(t, retryMaxAttempts, calls)
	})

	t.Run("stops when parent context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		err := withRetry(ctx, time.Second, func(context.Context) error {
			calls++
			cancel()
			return context.DeadlineExceeded
		})
		require.Error(t, err)
		require.Equal(t, 1, calls)
	})
}

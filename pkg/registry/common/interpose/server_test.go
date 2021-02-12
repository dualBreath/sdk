// Copyright (c) 2020-2021 Doc.ai and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package interpose_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/api/pkg/api/registry"

	"github.com/networkservicemesh/sdk/pkg/registry/common/interpose"
	"github.com/networkservicemesh/sdk/pkg/registry/core/next"
	"github.com/networkservicemesh/sdk/pkg/tools/stringurl"
)

const (
	nameSuffix     = "#interpose-nse"
	name           = "nse"
	validURL       = "tcp://0.0.0.0"
	commonResponse = "response"
)

var samples = []struct {
	name             string
	in, out          *registry.NetworkServiceEndpoint
	isInMap, failure bool
}{
	{
		name: "interpose NSE",
		in: &registry.NetworkServiceEndpoint{
			Name: name + nameSuffix,
			Url:  validURL,
		},
		out: &registry.NetworkServiceEndpoint{
			Name: name + nameSuffix,
			Url:  validURL,
		},
		isInMap: true,
	},
	{
		name: "common NSE",
		in: &registry.NetworkServiceEndpoint{
			Name: name,
		},
		out: &registry.NetworkServiceEndpoint{
			Name: commonResponse,
		},
	},
	{
		name: "invalid NSE",
		in: &registry.NetworkServiceEndpoint{
			Name: name + nameSuffix,
		},
		isInMap: false,
		failure: true,
	},
}

func TestInterposeRegistryServer(t *testing.T) {
	for i := range samples {
		sample := samples[i]
		t.Run(sample.name, func(t *testing.T) {
			var crossMap stringurl.Map
			server := next.NewNetworkServiceEndpointRegistryServer(
				interpose.NewNetworkServiceEndpointRegistryServer(&crossMap),
				new(testRegistry),
			)

			reg, err := server.Register(context.Background(), sample.in)
			if sample.failure {
				require.Error(t, err)

				requireCrossMapEqual(t, map[string]string{}, &crossMap)
			} else {
				require.NoError(t, err)
				require.Equal(t, sample.out.String(), reg.String())

				if sample.isInMap {
					requireCrossMapEqual(t, map[string]string{
						sample.in.Name: sample.in.Url,
					}, &crossMap)
				} else {
					requireCrossMapEqual(t, map[string]string{}, &crossMap)
				}

				_, err := server.Unregister(context.Background(), reg)
				require.NoError(t, err)

				requireCrossMapEqual(t, map[string]string{}, &crossMap)
			}
		})
	}
}

func requireCrossMapEqual(t *testing.T, expected map[string]string, crossMap *stringurl.Map) {
	actual := map[string]string{}
	crossMap.Range(func(key string, value *url.URL) bool {
		actual[key] = value.String()
		return true
	})
	require.Equal(t, expected, actual)
}

type testRegistry struct {
	registry.NetworkServiceEndpointRegistryServer
}

func (r *testRegistry) Register(ctx context.Context, in *registry.NetworkServiceEndpoint) (*registry.NetworkServiceEndpoint, error) {
	in.Name = commonResponse
	return next.NetworkServiceEndpointRegistryServer(ctx).Register(ctx, in)
}

func (r *testRegistry) Unregister(ctx context.Context, in *registry.NetworkServiceEndpoint) (*empty.Empty, error) {
	return next.NetworkServiceEndpointRegistryServer(ctx).Unregister(ctx, in)
}

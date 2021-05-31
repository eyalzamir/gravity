/*
Copyright 2018 Gravitational, Inc.

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

package client

import (
	"context"
	"time"

	pb "github.com/gravitational/gravity/lib/rpc/proto"
	"github.com/gravitational/gravity/lib/storage"

	"github.com/gogo/protobuf/types"
	"github.com/gravitational/trace"
	"github.com/gravitational/trace/trail"
)

// GetSystemInfo queries remote system information
func (c *Client) GetSystemInfo(ctx context.Context) (storage.System, error) {
	resp, err := c.discovery.GetSystemInfo(ctx, &types.Empty{})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	system, err := storage.UnmarshalSystemInfo(resp.Payload)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return system, nil
}

// GetRuntimeConfig returns agent's runtime configuration
func (c *Client) GetRuntimeConfig(ctx context.Context) (*pb.RuntimeConfig, error) {
	config, err := c.discovery.GetRuntimeConfig(ctx, &types.Empty{})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return config, nil
}

// GetCurrentTime returns agent's current time as UTC timestamp
func (c *Client) GetCurrentTime(ctx context.Context) (*time.Time, error) {
	proto, err := c.discovery.GetCurrentTime(ctx, &types.Empty{})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	ts, err := types.TimestampFromProto(proto)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &ts, nil
}

// GetVersion returns agent's version information
func (c *Client) GetVersion(ctx context.Context) (*pb.Version, error) {
	version, err := c.discovery.GetVersion(ctx, &types.Empty{})
	if err != nil {
		return nil, trail.FromGRPC(err)
	}
	return version, nil
}

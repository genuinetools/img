package client

import (
	"context"
	"fmt"

	controlapi "github.com/moby/buildkit/api/services/control"
)

// DiskUsage returns the disk usage being consumed by the buildkit controller.
func (c *Client) DiskUsage(ctx context.Context, req *controlapi.DiskUsageRequest) (*controlapi.DiskUsageResponse, error) {
	if c.controller == nil {
		// Create the controller.
		if err := c.createController(); err != nil {
			return nil, err
		}
	}

	// Call diskusage.
	resp, err := c.controller.DiskUsage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("getting disk usage failed: %v", err)
	}

	return resp, nil
}

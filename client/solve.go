package client

import (
	"context"
	"fmt"

	controlapi "github.com/moby/buildkit/api/services/control"
)

// Solve calls Solve on the controller.
func (c *Client) Solve(ctx context.Context, req *controlapi.SolveRequest) error {
	if c.controller == nil {
		// Create the controller.
		if err := c.createController(); err != nil {
			return err
		}
	}

	// Call solve.
	if _, err := c.controller.Solve(ctx, req); err != nil {
		return fmt.Errorf("solving failed: %v", err)
	}

	return nil
}

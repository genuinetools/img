package client

import (
	"context"
	"fmt"

	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/worker/base"
	"golang.org/x/sync/errgroup"
)

// Prune calls Prune on the worker.
func (c *Client) Prune(ctx context.Context) ([]*controlapi.UsageRecord, error) {
	ch := make(chan client.UsageInfo)

	// Create the worker opts.
	opt, err := c.createWorkerOpt(false)
	if err != nil {
		return nil, fmt.Errorf("creating worker opt failed: %v", err)
	}

	// Create the new worker.
	w, err := base.NewWorker(opt)
	if err != nil {
		return nil, fmt.Errorf("creating worker failed: %v", err)
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		// Call prune on the worker.
		return w.Prune(ctx, ch)
	})

	eg2, ctx := errgroup.WithContext(ctx)
	eg2.Go(func() error {
		defer close(ch)
		return eg.Wait()
	})

	usage := []*controlapi.UsageRecord{}
	eg2.Go(func() error {
		for r := range ch {
			usage = append(usage, &controlapi.UsageRecord{
				ID:          r.ID,
				Mutable:     r.Mutable,
				InUse:       r.InUse,
				Size_:       r.Size,
				Parent:      r.Parent,
				UsageCount:  int64(r.UsageCount),
				Description: r.Description,
				CreatedAt:   r.CreatedAt,
				LastUsedAt:  r.LastUsedAt,
			})
		}

		return nil
	})

	return usage, eg2.Wait()
}

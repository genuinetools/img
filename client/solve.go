package client

import (
	"context"
	"time"

	controlapi "github.com/moby/buildkit/api/services/control"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

// Solve calls Solve on the controller.
func (c *Client) Solve(ctx context.Context, req *controlapi.SolveRequest, ch chan *controlapi.StatusResponse) error {
	defer close(ch)
	if c.controller == nil {
		// Create the controller.
		if err := c.createController(); err != nil {
			return err
		}
	}

	statusCtx, cancelStatus := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer func() { // make sure the Status ends cleanly on build errors
			go func() {
				<-time.After(3 * time.Second)
				cancelStatus()
			}()
		}()
		_, err := c.controller.Solve(ctx, req)
		if err != nil {
			return errors.Wrap(err, "failed to solve")
		}
		return nil
	})

	eg.Go(func() error {
		srv := &controlStatusServer{
			ctx: statusCtx,
			ch:  ch,
		}
		return c.controller.Status(&controlapi.StatusRequest{
			Ref: req.Ref,
		}, srv)
	})
	return eg.Wait()
}

type controlStatusServer struct {
	ctx               context.Context
	ch                chan *controlapi.StatusResponse
	grpc.ServerStream // dummy
}

func (x *controlStatusServer) SendMsg(m interface{}) error {
	return x.Send(m.(*controlapi.StatusResponse))
}

func (x *controlStatusServer) Send(m *controlapi.StatusResponse) error {
	x.ch <- m
	return nil
}

func (x *controlStatusServer) Context() context.Context {
	return x.ctx
}

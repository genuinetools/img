package client

import (
	"fmt"

	"github.com/moby/buildkit/control"
	"github.com/moby/buildkit/frontend"
	"github.com/moby/buildkit/frontend/dockerfile"
	"github.com/moby/buildkit/worker"
)

func (c *Client) createController() error {
	// Create the worker opts.
	opt, err := c.createWorkerOpt()
	if err != nil {
		return fmt.Errorf("creating worker opt failed: %v", err)
	}

	// Create the new worker.
	w, err := c.newWorker(opt)
	if err != nil {
		return fmt.Errorf("creating worker failed: %v", err)
	}

	// Create the worker controller.
	wc := &worker.Controller{}
	if err := wc.Add(w); err != nil {
		return fmt.Errorf("adding worker to worker controller failed: %v", err)
	}

	// Add the frontends.
	frontends := map[string]frontend.Frontend{}
	frontends["dockerfile.v0"] = dockerfile.NewDockerfileFrontend()

	// Create the controller.
	controller, err := control.NewController(control.Opt{
		WorkerController: wc,
		Frontends:        frontends,
		CacheExporter:    w.CacheExporter,
		CacheImporter:    w.CacheImporter,
	})
	if err != nil {
		return fmt.Errorf("creating new controller failed: %v", err)
	}

	// Set the controller for the client.
	c.controller = controller

	return nil
}

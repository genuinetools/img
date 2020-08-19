package client

import (
	"fmt"
	"path/filepath"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/moby/buildkit/cache/remotecache"
	inlineremotecache "github.com/moby/buildkit/cache/remotecache/inline"
	localremotecache "github.com/moby/buildkit/cache/remotecache/local"
	registryremotecache "github.com/moby/buildkit/cache/remotecache/registry"
	"github.com/moby/buildkit/control"
	"github.com/moby/buildkit/frontend"
	"github.com/moby/buildkit/frontend/dockerfile/builder"
	"github.com/moby/buildkit/frontend/gateway"
	"github.com/moby/buildkit/frontend/gateway/forwarder"
	"github.com/moby/buildkit/solver/bboltcachestorage"
	"github.com/moby/buildkit/worker"
	"github.com/moby/buildkit/worker/base"
)

func (c *Client) createController() error {
	sm, err := c.getSessionManager()
	if err != nil {
		return fmt.Errorf("creating session manager failed: %v", err)
	}
	// Create the worker opts.
	opt, err := c.createWorkerOpt(true)
	if err != nil {
		return fmt.Errorf("creating worker opt failed: %v", err)
	}

	// Create the new worker.
	w, err := base.NewWorker(opt)
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
	frontends["dockerfile.v0"] = forwarder.NewGatewayForwarder(wc, builder.Build)
	frontends["gateway.v0"] = gateway.NewGatewayFrontend(wc)

	// Create the cache storage
	cacheStorage, err := bboltcachestorage.NewStore(filepath.Join(c.root, "cache.db"))
	if err != nil {
		return err
	}

	remoteCacheExporterFuncs := map[string]remotecache.ResolveCacheExporterFunc{
		"inline":   inlineremotecache.ResolveCacheExporterFunc(),
		"local":    localremotecache.ResolveCacheExporterFunc(sm),
		"registry": registryremotecache.ResolveCacheExporterFunc(sm, docker.ConfigureDefaultRegistries()),
	}
	remoteCacheImporterFuncs := map[string]remotecache.ResolveCacheImporterFunc{
		"local":    localremotecache.ResolveCacheImporterFunc(sm),
		"registry": registryremotecache.ResolveCacheImporterFunc(sm, opt.ContentStore, docker.ConfigureDefaultRegistries()),
	}

	// Create the controller.
	controller, err := control.NewController(control.Opt{
		SessionManager:            sm,
		WorkerController:          wc,
		Frontends:                 frontends,
		ResolveCacheExporterFuncs: remoteCacheExporterFuncs,
		ResolveCacheImporterFuncs: remoteCacheImporterFuncs,
		CacheKeyStorage:           cacheStorage,
	})
	if err != nil {
		return fmt.Errorf("creating new controller failed: %v", err)
	}

	// Set the controller for the client.
	c.controller = controller

	return nil
}

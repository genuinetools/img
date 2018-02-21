package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/jessfraz/img/worker/runc"
	"github.com/moby/buildkit/control"
	"github.com/moby/buildkit/frontend"
	"github.com/moby/buildkit/frontend/dockerfile"
	"github.com/moby/buildkit/worker"
	"github.com/sirupsen/logrus"
)

func createController(cmd command) (*control.Controller, *fuse.Server, error) {
	// Create the runc worker.
	opt, fuseserver, err := runc.NewWorkerOpt(stateDir, backend)
	if err != nil {
		return nil, fuseserver, fmt.Errorf("creating runc worker opt failed: %v", err)
	}

	localDirs := getLocalDirs(cmd)

	w, err := runc.NewWorker(opt, localDirs)
	if err != nil {
		return nil, fuseserver, err
	}

	// Create the worker controller.
	wc := &worker.Controller{}
	if err = wc.Add(w); err != nil {
		return nil, fuseserver, err
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
	return controller, fuseserver, err
}

func getLocalDirs(c command) map[string]string {
	// only return the local dirs for buildCommand
	cmd, ok := c.(*buildCommand)
	if !ok {
		return nil
	}

	file := cmd.dockerfilePath
	if file == "" {
		file = filepath.Join(cmd.contextDir, "Dockerfile")
	}

	return map[string]string{
		"context":    cmd.contextDir,
		"dockerfile": filepath.Dir(file),
	}
}

// On ^C, SIGTERM, etc handle exit.
func handleSignals(fuseserver *fuse.Server) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT)
	go func() {
		for sig := range ch {
			if fuseserver == nil {
				logrus.Infof("Received %s, shutting down", sig.String())
				os.Exit(1)
			}

			if err := fuseserver.Unmount(); err != nil {
				logrus.Errorf("Unmounting FUSE server failed: %v", err)
			}
			logrus.Infof("Received %s, unmounting FUSE Server", sig.String())
			os.Exit(1)
		}
	}()
}

// Unmount the fuseserver.
func unmount(fuseserver *fuse.Server) {
	if fuseserver == nil {
		return
	}

	if err := fuseserver.Unmount(); err != nil {
		logrus.Errorf("Unmounting FUSE server failed: %v", err)
	}
}

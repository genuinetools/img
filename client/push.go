package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/docker/distribution/reference"
	"github.com/moby/buildkit/util/push"
)

// Push sends an image to a remote registry.
func (c *Client) Push(ctx context.Context, image string, insecure bool) error {
	// Parse the image name and tag.
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return fmt.Errorf("parsing image name %q failed: %v", image, err)
	}
	// Add the latest lag if they did not provide one.
	named = reference.TagNameOnly(named)
	image = named.String()

	// Create the worker opts.
	opt, err := c.createWorkerOpt(false)
	if err != nil {
		return fmt.Errorf("creating worker opt failed: %v", err)
	}

	imgObj, err := opt.ImageStore.Get(ctx, image)
	if err != nil {
		return fmt.Errorf("getting image %q failed: %v", image, err)
	}

	sm, err := c.getSessionManager()
	if err != nil {
		return err
	}

	if err := push.Push(ctx, sm, opt.ContentStore, imgObj.Target.Digest, image, insecure, opt.RegistryHosts, false); err != nil {
		if !isErrHTTPResponseToHTTPSClient(err) {
			return err
		}

		if !insecure {
			return err
		}

		return push.Push(ctx, sm, opt.ContentStore, imgObj.Target.Digest, image, insecure, registryHostsWithPlainHTTP(), false)
	}
	return nil
}

func isErrHTTPResponseToHTTPSClient(err error) bool {
	// The error string is unexposed as of Go 1.13, so we can't use `errors.Is`.
	// https://github.com/golang/go/issues/44855

	const unexposed = "server gave HTTP response to HTTPS client"
	return strings.Contains(err.Error(), unexposed)
}

func registryHostsWithPlainHTTP() docker.RegistryHosts {
	return docker.ConfigureDefaultRegistries(docker.WithPlainHTTP(func(_ string) (bool, error) {
		return true, nil
	}))
}

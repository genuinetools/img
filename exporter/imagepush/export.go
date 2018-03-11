package imagepush

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/images"
	"github.com/genuinetools/img/exporter/containerimage"
	"github.com/genuinetools/img/util/push"
	"github.com/moby/buildkit/cache"
	"github.com/moby/buildkit/exporter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	keyImageName = "name"
	keyInsecure  = "registry.insecure"
)

// Opt contains the options for the container image pusher.
type Opt struct {
	ImageWriter *containerimage.ImageWriter
	Images      images.Store
}

type imagePusher struct {
	opt Opt
}

// New returns a new container image exporter.
func New(opt Opt) (exporter.Exporter, error) {
	im := &imagePusher{opt: opt}
	return im, nil
}

// Resolve returns an exporter instance.
func (e *imagePusher) Resolve(ctx context.Context, opt map[string]string) (exporter.ExporterInstance, error) {
	i := &imagePusherInstance{imagePusher: e}
	for k, v := range opt {
		switch k {
		case keyImageName:
			i.targetName = v
		case keyInsecure:
			i.insecure = true
		default:
			logrus.Warnf("image pusher: unknown option %s", k)
		}
	}
	return i, nil
}

type imagePusherInstance struct {
	*imagePusher
	targetName string
	insecure   bool
}

func (e *imagePusherInstance) Name() string {
	return "pushing to registry"
}

// Export commits the image and pushes it to a registry if that option is passed.
func (e *imagePusherInstance) Export(ctx context.Context, ref cache.ImmutableRef, opt map[string][]byte) error {
	if e.targetName == "" {
		return errors.New("target name cannot be empty")
	}

	if e.opt.Images == nil {
		return errors.New("image store is nil")
	}

	image, err := e.opt.Images.Get(ctx, e.targetName)
	if err != nil {
		return fmt.Errorf("getting target %s from image store failed: %v", e.targetName, err)
	}

	return push.Push(ctx, e.opt.ImageWriter.ContentStore(), image.Target.Digest, e.targetName, e.insecure)
}

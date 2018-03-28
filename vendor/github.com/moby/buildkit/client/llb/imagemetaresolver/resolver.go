package imagemetaresolver

import (
	"context"
	"net/http"
	"sync"

	"github.com/BurntSushi/locker"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/util/contentutil"
	"github.com/moby/buildkit/util/imageutil"
	digest "github.com/opencontainers/go-digest"
)

var defaultImageMetaResolver llb.ImageMetaResolver
var defaultImageMetaResolverOnce sync.Once

var WithDefault = llb.ImageOptionFunc(func(ii *llb.ImageInfo) {
	llb.WithMetaResolver(Default()).SetImageOption(ii)
})

func New() llb.ImageMetaResolver {
	return &imageMetaResolver{
		resolver: docker.NewResolver(docker.ResolverOptions{
			Client: http.DefaultClient,
		}),
		buffer: contentutil.NewBuffer(),
		cache:  map[string]resolveResult{},
		locker: locker.NewLocker(),
	}
}

func Default() llb.ImageMetaResolver {
	defaultImageMetaResolverOnce.Do(func() {
		defaultImageMetaResolver = New()
	})
	return defaultImageMetaResolver
}

type imageMetaResolver struct {
	resolver remotes.Resolver
	buffer   contentutil.Buffer
	locker   *locker.Locker
	cache    map[string]resolveResult
}

type resolveResult struct {
	config []byte
	dgst   digest.Digest
}

func (imr *imageMetaResolver) ResolveImageConfig(ctx context.Context, ref string) (digest.Digest, []byte, error) {
	imr.locker.Lock(ref)
	defer imr.locker.Unlock(ref)

	if res, ok := imr.cache[ref]; ok {
		return res.dgst, res.config, nil
	}

	dgst, config, err := imageutil.Config(ctx, ref, imr.resolver, imr.buffer)
	if err != nil {
		return "", nil, err
	}

	imr.cache[ref] = resolveResult{dgst: dgst, config: config}
	return dgst, config, nil
}

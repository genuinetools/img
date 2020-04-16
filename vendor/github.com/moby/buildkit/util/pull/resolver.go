package pull

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	distreference "github.com/docker/distribution/reference"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/source"
	"github.com/moby/buildkit/util/resolver"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

var cache *resolverCache

func init() {
	cache = newResolverCache()
}

func NewResolver(ctx context.Context, hosts docker.RegistryHosts, sm *session.Manager, imageStore images.Store, mode source.ResolveMode, ref string) remotes.Resolver {
	if res := cache.Get(ctx, ref); res != nil {
		return withLocal(res, imageStore, mode)
	}

	r := resolver.New(ctx, hosts, sm)
	r = cache.Add(ctx, ref, r)

	return withLocal(r, imageStore, mode)
}

func EnsureManifestRequested(ctx context.Context, res remotes.Resolver, ref string) {
	rr := res
	lr, ok := res.(withLocalResolver)
	if ok {
		if atomic.LoadInt64(&lr.counter) > 0 {
			return
		}
		rr = lr.Resolver
	}
	cr, ok := rr.(*cachedResolver)
	if !ok {
		return
	}
	if atomic.LoadInt64(&cr.counter) == 0 {
		res.Resolve(ctx, ref)
	}
}

func withLocal(r remotes.Resolver, imageStore images.Store, mode source.ResolveMode) remotes.Resolver {
	if imageStore == nil || mode == source.ResolveModeForcePull {
		return r
	}

	return withLocalResolver{Resolver: r, is: imageStore, mode: mode}
}

// A remotes.Resolver which checks the local image store if the real
// resolver cannot find the image, essentially falling back to a local
// image if one is present.
//
// We do not override the Fetcher or Pusher methods:
//
// - Fetcher is called by github.com/containerd/containerd/remotes/:fetch()
//   only after it has checked for the content locally, so avoid the
//   hassle of interposing a local-fetch proxy and simply pass on the
//   request.
// - Pusher wouldn't make sense to push locally, so just forward.

type withLocalResolver struct {
	counter int64 // needs to be 64bit aligned for 32bit systems
	remotes.Resolver
	is   images.Store
	mode source.ResolveMode
}

func (r withLocalResolver) Resolve(ctx context.Context, ref string) (string, ocispec.Descriptor, error) {
	if r.mode == source.ResolveModePreferLocal {
		if img, err := r.is.Get(ctx, ref); err == nil {
			atomic.AddInt64(&r.counter, 1)
			return ref, img.Target, nil
		}
	}

	n, desc, err := r.Resolver.Resolve(ctx, ref)
	if err == nil {
		return n, desc, err
	}

	if r.mode == source.ResolveModeDefault {
		if img, err := r.is.Get(ctx, ref); err == nil {
			return ref, img.Target, nil
		}
	}

	return "", ocispec.Descriptor{}, err
}

type resolverCache struct {
	mu sync.Mutex
	m  map[string]cachedResolver
}

type cachedResolver struct {
	counter int64
	timeout time.Time
	remotes.Resolver
}

func (cr *cachedResolver) Resolve(ctx context.Context, ref string) (name string, desc ocispec.Descriptor, err error) {
	atomic.AddInt64(&cr.counter, 1)
	return cr.Resolver.Resolve(ctx, ref)
}

func (r *resolverCache) Add(ctx context.Context, ref string, resolver remotes.Resolver) remotes.Resolver {
	r.mu.Lock()
	defer r.mu.Unlock()

	ref = r.repo(ref) + "-" + session.FromContext(ctx)

	cr, ok := r.m[ref]
	cr.timeout = time.Now().Add(time.Minute)
	if ok {
		return &cr
	}

	cr.Resolver = resolver
	r.m[ref] = cr
	return &cr
}

func (r *resolverCache) repo(refStr string) string {
	ref, err := distreference.ParseNormalizedNamed(refStr)
	if err != nil {
		return refStr
	}
	return ref.Name()
}

func (r *resolverCache) Get(ctx context.Context, ref string) remotes.Resolver {
	r.mu.Lock()
	defer r.mu.Unlock()

	ref = r.repo(ref) + "-" + session.FromContext(ctx)

	cr, ok := r.m[ref]
	if !ok {
		return nil
	}
	return &cr
}

func (r *resolverCache) clean(now time.Time) {
	r.mu.Lock()
	for k, cr := range r.m {
		if now.After(cr.timeout) {
			delete(r.m, k)
		}
	}
	r.mu.Unlock()
}

func newResolverCache() *resolverCache {
	rc := &resolverCache{
		m: map[string]cachedResolver{},
	}
	t := time.NewTicker(time.Minute)
	go func() {
		for {
			rc.clean(<-t.C)
		}
	}()
	return rc
}

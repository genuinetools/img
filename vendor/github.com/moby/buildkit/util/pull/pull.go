package pull

import (
	"context"
	"sync"
	"time"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/diff"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/platforms"
	"github.com/containerd/containerd/reference"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/remotes/docker/schema1"
	"github.com/moby/buildkit/snapshot"
	"github.com/moby/buildkit/util/imageutil"
	"github.com/moby/buildkit/util/progress"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

type Puller struct {
	Snapshotter  snapshot.Snapshotter
	ContentStore content.Store
	Applier      diff.Applier
	Src          reference.Spec
	Platform     *ocispec.Platform
	// See NewResolver()
	Resolver    remotes.Resolver
	resolveOnce sync.Once
	desc        ocispec.Descriptor
	ref         string
	resolveErr  error
}

type Pulled struct {
	Ref           string
	Descriptor    ocispec.Descriptor
	Layers        []ocispec.Descriptor
	MetadataBlobs []ocispec.Descriptor
}

func (p *Puller) Resolve(ctx context.Context) (string, ocispec.Descriptor, error) {
	p.resolveOnce.Do(func() {
		resolveProgressDone := oneOffProgress(ctx, "resolve "+p.Src.String())

		desc := ocispec.Descriptor{
			Digest: p.Src.Digest(),
		}
		if desc.Digest != "" {
			info, err := p.ContentStore.Info(ctx, desc.Digest)
			if err == nil {
				desc.Size = info.Size
				p.ref = p.Src.String()
				ra, err := p.ContentStore.ReaderAt(ctx, desc)
				if err == nil {
					mt, err := imageutil.DetectManifestMediaType(ra)
					if err == nil {
						desc.MediaType = mt
						p.desc = desc
						resolveProgressDone(nil)
						return
					}
				}
			}
		}

		ref, desc, err := p.Resolver.Resolve(ctx, p.Src.String())
		if err != nil {
			p.resolveErr = err
			resolveProgressDone(err)
			return
		}
		p.desc = desc
		p.ref = ref
		resolveProgressDone(nil)
	})
	return p.ref, p.desc, p.resolveErr
}

func (p *Puller) Pull(ctx context.Context) (*Pulled, error) {
	if _, _, err := p.Resolve(ctx); err != nil {
		return nil, err
	}

	var platform platforms.MatchComparer
	if p.Platform != nil {
		platform = platforms.Only(*p.Platform)
	} else {
		platform = platforms.Default()
	}

	ongoing := newJobs(p.ref)

	pctx, stopProgress := context.WithCancel(ctx)

	go showProgress(pctx, ongoing, p.ContentStore)

	fetcher, err := p.Resolver.Fetcher(ctx, p.ref)
	if err != nil {
		stopProgress()
		return nil, err
	}

	// TODO: need a wrapper snapshot interface that combines content
	// and snapshots as 1) buildkit shouldn't have a dependency on contentstore
	// or 2) cachemanager should manage the contentstore
	handlers := []images.Handler{
		images.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
			ongoing.add(desc)
			return nil, nil
		}),
	}
	var schema1Converter *schema1.Converter
	if p.desc.MediaType == images.MediaTypeDockerSchema1Manifest {
		schema1Converter = schema1.NewConverter(p.ContentStore, fetcher)
		handlers = append(handlers, schema1Converter)
	} else {
		// Get all the children for a descriptor
		childrenHandler := images.ChildrenHandler(p.ContentStore)
		// Filter the children by the platform
		childrenHandler = images.FilterPlatforms(childrenHandler, platform)
		// Limit manifests pulled to the best match in an index
		childrenHandler = images.LimitManifests(childrenHandler, platform, 1)

		dslHandler, err := docker.AppendDistributionSourceLabel(p.ContentStore, p.ref)
		if err != nil {
			stopProgress()
			return nil, err
		}
		handlers = append(handlers,
			remotes.FetchHandler(p.ContentStore, fetcher),
			childrenHandler,
			dslHandler,
		)
	}

	if err := images.Dispatch(ctx, images.Handlers(handlers...), nil, p.desc); err != nil {
		stopProgress()
		return nil, err
	}
	stopProgress()

	var usedBlobs, unusedBlobs []ocispec.Descriptor

	if schema1Converter != nil {
		ongoing.remove(p.desc) // Not left in the content store so this is sufficient.
		p.desc, err = schema1Converter.Convert(ctx)
		if err != nil {
			return nil, err
		}
		ongoing.add(p.desc)

		var mu sync.Mutex // images.Dispatch calls handlers in parallel
		allBlobs := make(map[digest.Digest]ocispec.Descriptor)
		for _, j := range ongoing.added {
			allBlobs[j.Digest] = j.Descriptor
		}

		handlers := []images.Handler{
			images.HandlerFunc(func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
				mu.Lock()
				defer mu.Unlock()
				usedBlobs = append(usedBlobs, desc)
				delete(allBlobs, desc.Digest)
				return nil, nil
			}),
			images.FilterPlatforms(images.ChildrenHandler(p.ContentStore), platform),
		}

		if err := images.Dispatch(ctx, images.Handlers(handlers...), nil, p.desc); err != nil {
			return nil, err
		}

		for _, j := range allBlobs {
			unusedBlobs = append(unusedBlobs, j)
		}
	} else {
		for _, j := range ongoing.added {
			usedBlobs = append(usedBlobs, j.Descriptor)
		}
	}

	// split all pulled data to layers and rest. layers remain roots and are deleted with snapshots. rest will be linked to layers.
	var notLayerBlobs []ocispec.Descriptor
	var layerBlobs []ocispec.Descriptor
	for _, j := range usedBlobs {
		switch j.MediaType {
		case ocispec.MediaTypeImageLayer, images.MediaTypeDockerSchema2Layer, ocispec.MediaTypeImageLayerGzip, images.MediaTypeDockerSchema2LayerGzip, images.MediaTypeDockerSchema2LayerForeign, images.MediaTypeDockerSchema2LayerForeignGzip:
			layerBlobs = append(layerBlobs, j)
		default:
			notLayerBlobs = append(notLayerBlobs, j)
		}
	}

	layers, err := getLayers(ctx, p.ContentStore, p.desc, platform)
	if err != nil {
		return nil, err
	}

	return &Pulled{
		Ref:           p.ref,
		Descriptor:    p.desc,
		Layers:        layers,
		MetadataBlobs: notLayerBlobs,
	}, nil
}

func getLayers(ctx context.Context, provider content.Provider, desc ocispec.Descriptor, platform platforms.MatchComparer) ([]ocispec.Descriptor, error) {
	manifest, err := images.Manifest(ctx, provider, desc, platform)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	image := images.Image{Target: desc}
	diffIDs, err := image.RootFS(ctx, provider, platform)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve rootfs")
	}
	if len(diffIDs) != len(manifest.Layers) {
		return nil, errors.Errorf("mismatched image rootfs and manifest layers %+v %+v", diffIDs, manifest.Layers)
	}
	layers := make([]ocispec.Descriptor, len(diffIDs))
	for i := range diffIDs {
		desc := manifest.Layers[i]
		if desc.Annotations == nil {
			desc.Annotations = map[string]string{}
		}
		desc.Annotations["containerd.io/uncompressed"] = diffIDs[i].String()
		layers[i] = desc
	}
	return layers, nil
}

func showProgress(ctx context.Context, ongoing *jobs, cs content.Store) {
	var (
		ticker   = time.NewTicker(150 * time.Millisecond)
		statuses = map[string]statusInfo{}
		done     bool
	)
	defer ticker.Stop()

	pw, _, ctx := progress.FromContext(ctx)
	defer pw.Close()

	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			done = true
		}

		resolved := "resolved"
		if !ongoing.isResolved() {
			resolved = "resolving"
		}
		statuses[ongoing.name] = statusInfo{
			Ref:    ongoing.name,
			Status: resolved,
		}

		actives := make(map[string]statusInfo)

		if !done {
			active, err := cs.ListStatuses(ctx, "")
			if err != nil {
				// log.G(ctx).WithError(err).Error("active check failed")
				continue
			}
			// update status of active entries!
			for _, active := range active {
				actives[active.Ref] = statusInfo{
					Ref:       active.Ref,
					Status:    "downloading",
					Offset:    active.Offset,
					Total:     active.Total,
					StartedAt: active.StartedAt,
					UpdatedAt: active.UpdatedAt,
				}
			}
		}

		// now, update the items in jobs that are not in active
		for _, j := range ongoing.jobs() {
			refKey := remotes.MakeRefKey(ctx, j.Descriptor)
			if a, ok := actives[refKey]; ok {
				started := j.started
				pw.Write(j.Digest.String(), progress.Status{
					Action:  a.Status,
					Total:   int(a.Total),
					Current: int(a.Offset),
					Started: &started,
				})
				continue
			}

			if !j.done {
				info, err := cs.Info(context.TODO(), j.Digest)
				if err != nil {
					if errdefs.IsNotFound(err) {
						pw.Write(j.Digest.String(), progress.Status{
							Action: "waiting",
						})
						continue
					}
				} else {
					j.done = true
				}

				if done || j.done {
					started := j.started
					createdAt := info.CreatedAt
					pw.Write(j.Digest.String(), progress.Status{
						Action:    "done",
						Current:   int(info.Size),
						Total:     int(info.Size),
						Completed: &createdAt,
						Started:   &started,
					})
				}
			}
		}
		if done {
			return
		}
	}
}

// jobs provides a way of identifying the download keys for a particular task
// encountering during the pull walk.
//
// This is very minimal and will probably be replaced with something more
// featured.
type jobs struct {
	name     string
	added    map[digest.Digest]*job
	mu       sync.Mutex
	resolved bool
}

type job struct {
	ocispec.Descriptor
	done    bool
	started time.Time
}

func newJobs(name string) *jobs {
	return &jobs{
		name:  name,
		added: make(map[digest.Digest]*job),
	}
}

func (j *jobs) add(desc ocispec.Descriptor) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if _, ok := j.added[desc.Digest]; ok {
		return
	}
	j.added[desc.Digest] = &job{
		Descriptor: desc,
		started:    time.Now(),
	}
}

func (j *jobs) remove(desc ocispec.Descriptor) {
	j.mu.Lock()
	defer j.mu.Unlock()

	delete(j.added, desc.Digest)
}

func (j *jobs) jobs() []*job {
	j.mu.Lock()
	defer j.mu.Unlock()

	descs := make([]*job, 0, len(j.added))
	for _, j := range j.added {
		descs = append(descs, j)
	}
	return descs
}

func (j *jobs) isResolved() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.resolved
}

type statusInfo struct {
	Ref       string
	Status    string
	Offset    int64
	Total     int64
	StartedAt time.Time
	UpdatedAt time.Time
}

func oneOffProgress(ctx context.Context, id string) func(err error) error {
	pw, _, _ := progress.FromContext(ctx)
	now := time.Now()
	st := progress.Status{
		Started: &now,
	}
	pw.Write(id, st)
	return func(err error) error {
		// TODO: set error on status
		now := time.Now()
		st.Completed = &now
		pw.Write(id, st)
		pw.Close()
		return err
	}
}

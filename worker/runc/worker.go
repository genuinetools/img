package runc

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	imageexporter "github.com/jessfraz/img/exporter/containerimage"
	"github.com/jessfraz/img/source/containerimage"
	"github.com/jessfraz/img/source/local"
	"github.com/moby/buildkit/cache"
	"github.com/moby/buildkit/cache/cacheimport"
	"github.com/moby/buildkit/cache/instructioncache"
	localcache "github.com/moby/buildkit/cache/instructioncache/local"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/executor"
	"github.com/moby/buildkit/exporter"
	localexporter "github.com/moby/buildkit/exporter/local"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/solver"
	"github.com/moby/buildkit/solver/llbop"
	"github.com/moby/buildkit/solver/pb"
	"github.com/moby/buildkit/source"
	"github.com/moby/buildkit/source/git"
	"github.com/moby/buildkit/source/http"
	"github.com/moby/buildkit/worker"
	"github.com/moby/buildkit/worker/base"
	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

// Worker is a local worker instance with dedicated snapshotter, cache, and so on.
type Worker struct {
	base.WorkerOpt
	CacheManager  cache.Manager
	SourceManager *source.Manager
	cache         instructioncache.InstructionCache
	Exporters     map[string]exporter.Exporter
	ImageSource   source.Source
	CacheExporter *cacheimport.CacheExporter
	CacheImporter *cacheimport.CacheImporter
}

// NewWorker instantiates a local worker.
func NewWorker(opt base.WorkerOpt, localDirs map[string]string) (*Worker, error) {
	cm, err := cache.NewManager(cache.ManagerOpt{
		Snapshotter:   opt.Snapshotter,
		MetadataStore: opt.MetadataStore,
	})
	if err != nil {
		return nil, err
	}

	ic := &localcache.LocalStore{
		MetadataStore: opt.MetadataStore,
		Cache:         cm,
	}

	sm, err := source.NewManager()
	if err != nil {
		return nil, err
	}

	is, err := containerimage.NewSource(containerimage.SourceOpt{
		Snapshotter:   opt.Snapshotter,
		ContentStore:  opt.ContentStore,
		Applier:       opt.Applier,
		CacheAccessor: cm,
		Images:        opt.ImageStore,
	})
	if err != nil {
		return nil, err
	}

	sm.Register(is)

	gs, err := git.NewSource(git.Opt{
		CacheAccessor: cm,
		MetadataStore: opt.MetadataStore,
	})
	if err != nil {
		return nil, err
	}

	sm.Register(gs)

	hs, err := http.NewSource(http.Opt{
		CacheAccessor: cm,
		MetadataStore: opt.MetadataStore,
	})
	if err != nil {
		return nil, err
	}

	sm.Register(hs)

	ss, err := local.NewSource(local.Opt{
		CacheAccessor: cm,
		MetadataStore: opt.MetadataStore,
		LocalDirs:     localDirs,
	})
	if err != nil {
		return nil, err
	}
	sm.Register(ss)

	exporters := map[string]exporter.Exporter{}

	iw, err := imageexporter.NewImageWriter(imageexporter.WriterOpt{
		Snapshotter:  opt.Snapshotter,
		ContentStore: opt.ContentStore,
		Differ:       opt.Differ,
	})
	if err != nil {
		return nil, err
	}

	imageExporter, err := imageexporter.New(imageexporter.Opt{
		Images:      opt.ImageStore,
		ImageWriter: iw,
	})
	if err != nil {
		return nil, err
	}
	exporters[client.ExporterImage] = imageExporter

	localExporter, err := localexporter.New(localexporter.Opt{})
	if err != nil {
		return nil, err
	}
	exporters[client.ExporterLocal] = localExporter

	ce := cacheimport.NewCacheExporter(cacheimport.ExporterOpt{
		Snapshotter:  opt.Snapshotter,
		ContentStore: opt.ContentStore,
		Differ:       opt.Differ,
	})

	ci := cacheimport.NewCacheImporter(cacheimport.ImportOpt{
		Snapshotter:   opt.Snapshotter,
		ContentStore:  opt.ContentStore,
		Applier:       opt.Applier,
		CacheAccessor: cm,
	})

	return &Worker{
		WorkerOpt:     opt,
		CacheManager:  cm,
		SourceManager: sm,
		cache:         ic,
		Exporters:     exporters,
		ImageSource:   is,
		CacheExporter: ce,
		CacheImporter: ci,
	}, nil
}

// ID returns the worker ID.
func (w *Worker) ID() string {
	return w.WorkerOpt.ID
}

// Labels returns the worker labels.
func (w *Worker) Labels() map[string]string {
	return w.WorkerOpt.Labels
}

// ResolveOp resolves a vertex returning an option.
func (w *Worker) ResolveOp(v solver.Vertex, s worker.SubBuilder) (solver.Op, error) {
	switch op := v.Sys().(type) {
	case *pb.Op_Source:
		return llbop.NewSourceOp(v, op, w.SourceManager)
	case *pb.Op_Exec:
		return llbop.NewExecOp(v, op, w.CacheManager, w.Executor)
	case *pb.Op_Build:
		return llbop.NewBuildOp(v, op, s)
	default:
		return nil, errors.Errorf("could not resolve %v", v)
	}
}

// ResolveImageConfig resolves an image config.
func (w *Worker) ResolveImageConfig(ctx context.Context, ref string) (digest.Digest, []byte, error) {
	// ImageSource is typically source/containerimage
	resolveImageConfig, ok := w.ImageSource.(resolveImageConfig)
	if !ok {
		return "", nil, errors.Errorf("worker %q does not implement ResolveImageConfig", w.ID())
	}
	return resolveImageConfig.ResolveImageConfig(ctx, ref)
}

type resolveImageConfig interface {
	ResolveImageConfig(ctx context.Context, ref string) (digest.Digest, []byte, error)
}

// Exec executes a worker.
func (w *Worker) Exec(ctx context.Context, meta executor.Meta, rootFS cache.ImmutableRef, stdin io.ReadCloser, stdout, stderr io.WriteCloser) error {
	active, err := w.CacheManager.New(ctx, rootFS)
	if err != nil {
		return err
	}
	defer active.Release(context.TODO())
	return w.Executor.Exec(ctx, meta, active, nil, stdin, stdout, stderr)
}

// DiskUsage returns the disk usage.
func (w *Worker) DiskUsage(ctx context.Context, opt client.DiskUsageInfo) ([]*client.UsageInfo, error) {
	return w.CacheManager.DiskUsage(ctx, opt)
}

// Prune cleans the cache.
func (w *Worker) Prune(ctx context.Context, ch chan client.UsageInfo) error {
	return w.CacheManager.Prune(ctx, ch)
}

// Exporter returns a given exporter that matches the name that is passed.
func (w *Worker) Exporter(name string) (exporter.Exporter, error) {
	exp, ok := w.Exporters[name]
	if !ok {
		return nil, errors.Errorf("exporter %q could not be found", name)
	}
	return exp, nil
}

// InstructionCache returns the cache.
func (w *Worker) InstructionCache() instructioncache.InstructionCache {
	return w.cache
}

// Labels is autility function to create the runtime specific label opject applied to a worker.
func Labels(executor, snapshotter string) map[string]string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	labels := map[string]string{
		worker.LabelOS:          runtime.GOOS,
		worker.LabelArch:        runtime.GOARCH,
		worker.LabelExecutor:    executor,
		worker.LabelSnapshotter: snapshotter,
		worker.LabelHostname:    hostname,
	}
	return labels
}

// ID reads the worker id from the `workerid` file.
// If it does not exist, it creates a random one,
func ID(root string) (string, error) {
	f := filepath.Join(root, "workerid")
	b, err := ioutil.ReadFile(f)
	if err != nil {
		if os.IsNotExist(err) {
			id := identity.NewID()
			err := ioutil.WriteFile(f, []byte(id), 0400)
			return id, err
		}

		return "", err
	}

	return string(b), nil
}

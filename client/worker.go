package client

import (
	imageexporter "github.com/genuinetools/img/exporter/containerimage"
	"github.com/genuinetools/img/source/containerimage"
	"github.com/genuinetools/img/source/local"
	"github.com/moby/buildkit/cache"
	"github.com/moby/buildkit/cache/cacheimport"
	localcache "github.com/moby/buildkit/cache/instructioncache/local"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/exporter"
	localexporter "github.com/moby/buildkit/exporter/local"
	"github.com/moby/buildkit/source"
	"github.com/moby/buildkit/source/git"
	"github.com/moby/buildkit/source/http"
	"github.com/moby/buildkit/worker/base"
)

// newWorker instantiates a local worker.
func (c *Client) newWorker(opt base.WorkerOpt) (*base.Worker, error) {
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
		LocalDirs:     c.localDirs,
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

	return &base.Worker{
		WorkerOpt:     opt,
		CacheManager:  cm,
		SourceManager: sm,
		Cache:         ic,
		Exporters:     exporters,
		ImageSource:   is,
		CacheExporter: ce,
		CacheImporter: ci,
	}, nil
}

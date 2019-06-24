module github.com/genuinetools/img

replace github.com/hashicorp/go-immutable-radix => github.com/tonistiigi/go-immutable-radix v0.0.0-20170803185627-826af9ccf0fe

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

require (
	github.com/containerd/console v0.0.0-20181022165439-0650fd9eeb50
	github.com/containerd/containerd v1.3.0-0.20190426060238-3a3f0aac8819
	github.com/containerd/go-runc v0.0.0-20180907222934-5a6d9f37cfa3
	github.com/cyphar/filepath-securejoin v0.2.2
	github.com/docker/cli v0.0.0-20190321234815-f40f9c240ab0
	github.com/docker/distribution v2.7.1-0.20190205005809-0d3efadf0154+incompatible
	github.com/docker/docker v1.14.0-0.20190319215453-e7b5f7dbe98c
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.3.3
	github.com/genuinetools/reg v0.16.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible // indirect
	github.com/gogo/googleapis v1.1.0 // indirect
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/moby/buildkit v0.5.1
	github.com/opencontainers/image-spec v1.0.1
	github.com/opencontainers/runc v1.0.1-0.20190307181833-2b18fe1d885e
	github.com/opencontainers/runtime-spec v1.0.1 // indirect
	github.com/opentracing-contrib/go-stdlib v0.0.0-20180702182724-07a764486eb1 // indirect
	github.com/opentracing/opentracing-go v1.0.2 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.3.0
	github.com/spf13/cobra v0.0.5
	go.etcd.io/bbolt v1.3.2
	golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f
	google.golang.org/grpc v1.15.0
)

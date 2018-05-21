package types

const (
	// AutoBackend is automatically resolved into either overlayfs or native.
	// TODO: support fuse
	AutoBackend = "auto"
	// FUSEBackend defines the FUSE backend.
	FUSEBackend = "fuse"
	// NativeBackend defines the native backend.
	NativeBackend = "native"
	// OverlayFSBackend defines the overlayfs backend.
	OverlayFSBackend = "overlayfs"
)

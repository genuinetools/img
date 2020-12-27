package types

const (
	// AutoBackend is automatically resolved into a backend based on what the
	// current system supports.
	AutoBackend = "auto"
	// NativeBackend defines the native backend.
	NativeBackend = "native"
	// OverlayFSBackend defines the overlayfs backend.
	OverlayFSBackend = "overlayfs"
	// FUSEOverlayFSBackend defines the fuse-overlayfs backend.
	FUSEOverlayFSBackend = "fuse-overlayfs"
)

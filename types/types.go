package types

const (
	// AutoBackend is automatically resolved into either overlayfs or native.
	AutoBackend = "auto"
	// NativeBackend defines the native backend.
	NativeBackend = "native"
	// OverlayFSBackend defines the overlayfs backend.
	OverlayFSBackend = "overlayfs"
)

package types

const (
	// FUSEBackend defines the FUSE backend.
	FUSEBackend = "fuse"
	// NaiveBackend defines the naive backend.
	NaiveBackend = "naive"
	// OverlayFSBackend defines the overlayfs backend.
	OverlayFSBackend = "overlayfs"
	// InUnshareEnv is the variable used to hold the enviornment for if
	// this has been unshared.
	InUnshareEnv = "IMG_IN_UNSHARE"
)

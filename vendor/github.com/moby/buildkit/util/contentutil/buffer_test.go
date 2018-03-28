package contentutil

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	digest "github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestReadWrite(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	b := NewBuffer()

	err := content.WriteBlob(ctx, b, "foo", bytes.NewBuffer([]byte("foo0")), -1, "")
	require.NoError(t, err)

	err = content.WriteBlob(ctx, b, "foo", bytes.NewBuffer([]byte("foo1")), 4, "")
	require.NoError(t, err)

	err = content.WriteBlob(ctx, b, "foo", bytes.NewBuffer([]byte("foo2")), 3, "")
	require.Error(t, err)

	err = content.WriteBlob(ctx, b, "foo", bytes.NewBuffer([]byte("foo3")), -1, digest.FromBytes([]byte("foo4")))
	require.Error(t, err)

	err = content.WriteBlob(ctx, b, "foo", bytes.NewBuffer([]byte("foo4")), -1, digest.FromBytes([]byte("foo4")))
	require.NoError(t, err)

	dt, err := content.ReadBlob(ctx, b, digest.FromBytes([]byte("foo1")))
	require.NoError(t, err)
	require.Equal(t, string(dt), "foo1")

	_, err = content.ReadBlob(ctx, b, digest.FromBytes([]byte("foo3")))
	require.Error(t, err)
	require.Equal(t, errors.Cause(err), errdefs.ErrNotFound)
}

func TestReaderAt(t *testing.T) {
	t.Parallel()
	ctx := context.TODO()

	b := NewBuffer()

	err := content.WriteBlob(ctx, b, "foo", bytes.NewBuffer([]byte("foobar")), -1, "")
	require.NoError(t, err)

	rdr, err := b.ReaderAt(ctx, digest.FromBytes([]byte("foobar")))
	require.NoError(t, err)

	require.Equal(t, int64(6), rdr.Size())

	buf := make([]byte, 3)

	n, err := rdr.ReadAt(buf, 1)
	require.NoError(t, err)
	require.Equal(t, "oob", string(buf[:n]))

	buf = make([]byte, 7)

	n, err = rdr.ReadAt(buf, 3)
	require.Error(t, err)
	require.Equal(t, err, io.EOF)
	require.Equal(t, "bar", string(buf[:n]))
}

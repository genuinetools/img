package main

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func extractTarFile(path string, r io.Reader) error {
	cmd := exec.Command("tar", "-x", "-C", path) // may need some extra options for users/permissions
	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func writeTarFile(path string, h *tar.Header, r io.Reader) error {
	target := filepath.Join(path, h.Name)

	switch h.Typeflag {
	case tar.TypeReg, tar.TypeRegA:
		f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(f, r); err != nil {
			return err
		}

		f.Chmod(h.FileInfo().Mode())
		// fp.Chmod(h.Uid, h.Gid)
		os.Chtimes(target, h.AccessTime, h.ModTime)
	case tar.TypeDir:
		if err := os.MkdirAll(target, h.FileInfo().Mode()); err != nil {
			return err
		}
	default:
		fmt.Printf("skip %q %v -> %v", h.Typeflag, h.Name, h.Linkname)
		return fmt.Errorf("unsupported file: %v", h)
	}

	return nil
}

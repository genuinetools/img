package ioutils

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Copy copies src to dest, doesn't matter if src is a directory or a file
func Copy(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return cp(src, dest, info)
}

func cp(src, dest string, info os.FileInfo) error {
	if info == nil {
		return errors.New("os.FileInfo cannot be nil and passed to cp")
	}

	if info.IsDir() {
		return copyDir(src, dest, info)
	}

	return copyFile(src, dest, info)
}

func copyFile(src, dest string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		// We got a symlink, skip it.
		// TODO(jessfraz): find a better way to do this.
		fmt.Printf("symlink: %s -> %s\n", src, dest)
		return nil
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

func copyDir(src, dest string, info os.FileInfo) error {
	if err := os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}

	infos, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}

	for _, info := range infos {
		if err := cp(filepath.Join(src, info.Name()), filepath.Join(dest, info.Name()), info); err != nil {
			return err
		}
	}

	return nil
}

package fsutils

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// CopieFile copies a file source to destination.
func CopyFile(source string, dest string) error {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()

	if _, err = io.Copy(df, sf); err != nil {
		return err
	}

	si, err := os.Stat(source)
	if err != nil {
		return err
	}

	return os.Chmod(dest, si.Mode())
}

// CopyDIr recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
func CopyDir(source string, dest string) error {
	// Get the properties of the source directory.
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return errors.New("CopyDir: Source is not a directory")
	}

	if _, err = os.Open(dest); !os.IsNotExist(err) && !DirIsEmpty(dest) {
		return errors.New("CopyDir: Destination already exists")
	}

	// Create the destination directory
	if err = os.MkdirAll(dest, fi.Mode()); err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sfp := source + "/" + entry.Name()
		dfp := dest + "/" + entry.Name()
		if entry.IsDir() {
			err = CopyDir(sfp, dfp)
			if err != nil {
				log.Println("copyDir:", err)
			}
		} else {
			// perform copy
			err = CopyFile(sfp, dfp)
			if err != nil {
				log.Println("copyDir:", err)
			}
		}

	}

	return nil
}

// DirIsEmpty checks if the directory is empty.
func DirIsEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()

	if _, err = f.Readdir(1); err == io.EOF {
		return true
	}

	return false
}

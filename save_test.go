package main

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveImage(t *testing.T) {
	runBuild(t, "savething", withDockerfile(`
    FROM busybox
	RUN echo savetest
    `))

	tmpf := filepath.Join(os.TempDir(), "save-image-test.tar")
	defer os.RemoveAll(tmpf)

	run(t, "save", "-o", tmpf, "savething")

	// Make sure the file exists
	if _, err := os.Stat(tmpf); os.IsNotExist(err) {
		t.Fatalf("%s should exist after saving the image but it didn't", tmpf)
	}
}

func TestSaveImageOCI(t *testing.T) {
	runBuild(t, "savethingoci", withDockerfile(`
    FROM busybox
	RUN echo savetest
    `))

	tmpf := filepath.Join(os.TempDir(), "save-oci-test.tar")
	defer os.RemoveAll(tmpf)

	run(t, "save", "--format", "oci", "-o", tmpf, "savethingoci")

	// Make sure the file exists
	if _, err := os.Stat(tmpf); os.IsNotExist(err) {
		t.Fatalf("%s should exist after saving the image but it didn't", tmpf)
	}
}

func TestSaveImageInvalid(t *testing.T) {
	runBuild(t, "savethinginvalid", withDockerfile(`
    FROM busybox
	RUN echo savetest
    `))

	tmpf := filepath.Join(os.TempDir(), "save-invalid.tar")
	defer os.RemoveAll(tmpf)

	out, err := doRun([]string{"save", "--format", "blah", "-o", tmpf, "savethinginvalid"}, nil)
	if err == nil {
		t.Fatalf("expected invalid format to fail but did not: %s", string(out))
	}
}

func TestSaveMultipleImages(t *testing.T) {
	var cases = []struct {
		format string
	}{
		{
			"",
		},
		{
			"docker",
		},
		{
			"oci",
		},
	}

	for _, tt := range cases {
		testname := tt.format
		t.Run(testname, func(t *testing.T) {

			runBuild(t, "multiimage1", withDockerfile(`
			FROM busybox
			RUN echo multiimage1
			`))

			runBuild(t, "multiimage2", withDockerfile(`
			FROM busybox
			RUN echo multiimage2
			`))

			tmpf := filepath.Join(os.TempDir(), fmt.Sprintf("save-multiple-%s.tar", tt.format))
			defer os.RemoveAll(tmpf)

			if tt.format != "" {
				run(t, "save", "--format", tt.format, "-o", tmpf, "multiimage1", "multiimage2")
			} else {
				run(t, "save", "-o", tmpf, "multiimage1", "multiimage2")
			}

			// Make sure the file exists
			if _, err := os.Stat(tmpf); os.IsNotExist(err) {
				t.Fatalf("%s should exist after saving the image but it didn't", tmpf)
			}

			count, err := getImageCountInTarball(tmpf)

			if err != nil {
				t.Fatal(err)
			}

			if count != 2 {
				t.Fatalf("should have 2 images in archive but have %d", count)
			}
		})
	}
}

func getImageCountInTarball(tarpath string) (int, error) {
	file, err := os.Open(tarpath)

	if err != nil {
		return -1, err
	}

	defer file.Close()

	tr := tar.NewReader(file)

	for {
		header, err := tr.Next()

		if err == io.EOF {
			return -1, errors.New("did not find manifest in tarball")
		}
		if err != nil {
			return -1, err
		}

		if header.Name == "manifest.json" {
			jsonFile, err := ioutil.ReadAll(tr)

			if err != nil {
				return -1, err
			}

			var result []map[string]string
			json.Unmarshal([]byte(jsonFile), &result)

			return len(result), nil
		}
	}
}

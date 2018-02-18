package runc

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/worker"
)

// BinaryExists checks if the runc binary exists.
func BinaryExists() bool {
	_, err := exec.LookPath("runc")
	if err != nil {
		return false
	}
	return true
}

// getOrCreateWorkerID reads the worker ID from the `workerid` file.
// If it does not exist, it creates a random one,
func getOrCreateWorkerID(root string) (string, error) {
	f := filepath.Join(root, "workerid")
	b, err := ioutil.ReadFile(f)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the id file.
			id := identity.NewID()

			err = ioutil.WriteFile(f, []byte(id), 0400)
			return id, err
		}

		return "", err
	}

	return string(b), nil
}

// createLabelsMap returns the relevant labels for this operating system.
func createLabelsMap(executor, snapshotter string) map[string]string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	labels := map[string]string{
		worker.LabelOS:          runtime.GOOS,
		worker.LabelArch:        runtime.GOARCH,
		worker.LabelExecutor:    executor,
		worker.LabelSnapshotter: snapshotter,
		worker.LabelHostname:    hostname,
	}

	return labels
}

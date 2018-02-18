package runc

import (
	"os"
	"testing"
)

func TestTempConsole(t *testing.T) {
	c, err := NewTempConsoleSocket()
	if err != nil {
		t.Fatal(err)
	}
	path := c.Path()
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatal("path still exists")
	}
}

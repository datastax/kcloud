package main

import (
	"runtime"
	"testing"
)

func TestRunCommand(t *testing.T) {
	var err error
	fooCmd := "foo"
	err = RunCommand(fooCmd)
	if err == nil {
		t.Fatal("expected foo to fail but it succeeded")
	}

	emptyCmd := "   "
	err = RunCommand(emptyCmd)
	if err == nil {
		t.Fatal("expected foo to fail but it succeeded")
	}

	gcloudCmd := "ls -l"
	if runtime.GOOS == "windows" {
		gcloudCmd = "dir"
	}
	err = RunCommand(gcloudCmd)
	if err != nil {
		t.Fatal(err)
	}
}

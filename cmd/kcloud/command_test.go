package main

import (
	"runtime"
	"testing"
)

func TestRunCommand(t *testing.T) {
	var err error
	fooCmd := "foo"
	err = RunCommandAndPrint(fooCmd)
	if err == nil {
		t.Fatal("expected foo to fail but it succeeded")
	}

	emptyCmd := "   "
	err = RunCommandAndPrint(emptyCmd)
	if err == nil {
		t.Fatal("expected foo to fail but it succeeded")
	}

	listCmd := "ls"
	if runtime.GOOS == "windows" {
		listCmd = "dir"
	}
	err = RunCommandAndPrint(listCmd)
	if err != nil {
		t.Fatal(err)
	}
}

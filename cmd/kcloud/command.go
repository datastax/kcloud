package main

import (
	"fmt"
	"os/exec"
)

// RunCommand runs the given command and prints the output
func RunCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	fmt.Print(string(output))
	if err != nil {
		return fmt.Errorf("failed command: %v %w", cmd, err)
	}
	return nil
}

package main

import (
	"fmt"
	"os/exec"
)

// RunCommand runs the given command and prints the output
func RunCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	if cli.Verbose {
		PrintCommand(command, cmd.Args...)
	}
	output, err := cmd.CombinedOutput()
	fmt.Print(string(output))
	if err != nil {
		return fmt.Errorf("failed command: %v %w", cmd, err)
	}
	return nil
}

// PrintCommand prints the given command with quote around the args for convenience
func PrintCommand(command string, args ...string) {
	fmt.Printf("DEBUG: command: %s", command)
	for _, arg := range args {
		fmt.Printf(" \"%s\"", arg)
	}
	fmt.Println()
}

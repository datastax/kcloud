package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunCommandAndPrint runs the given command and prints the output
func RunCommandAndPrint(command string, args ...string) error {
	if cli.Verbose {
		fmt.Println(QuoteCommand(command, args...))
	}
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed command: %v %w", cmd, err)
	}
	return nil
}

// RunCommand runs the given command and args and returns the
// combined stdout/stderr output
func RunCommand(command string, args ...string) ([]byte, error) {
	if cli.Verbose {
		fmt.Println(QuoteCommand(command, args...))
	}
	cmd := exec.Command(command, args...)
	return cmd.CombinedOutput()
}

// QuoteCommand writes the given command with quotes around the args for convenience
func QuoteCommand(command string, args ...string) string {
	var sb strings.Builder
	sb.WriteString("DEBUG: " + command)
	for _, arg := range args {
		sb.WriteString(fmt.Sprintf(" \"%s\"", arg))
	}
	return sb.String()
}

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// RunCommandAndPrint runs the given command and prints the output
func RunCommandAndPrint(command string, args ...string) error {
	if cli.Verbose {
		fmt.Println("[debug] ", QuoteCommand(command, args...))
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
		fmt.Println("[debug] ", QuoteCommand(command, args...))
	}
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

// QuoteCommand concatenates the command and args with quotes around the args to avoid whitespace issues.
func QuoteCommand(command string, args ...string) string {
	var quotedArgs strings.Builder
	for _, arg := range args {
		quotedArgs.WriteString((" " + strconv.Quote(arg)))
	}
	return command + quotedArgs.String()
}

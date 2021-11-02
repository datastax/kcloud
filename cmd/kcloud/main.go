package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func usage() {
	cmd := filepath.Base(os.Args[0])
	fmt.Printf("%s - retrieve kubernetes configuration from cloud providers\n", cmd)
	fmt.Printf("Usage: %s [options] aws|azr|gcp [account] [cluster-name]\n", cmd)
}

// ErrorOp indicates an unexpected error.
type ErrorOp struct{ Err error }

func (op ErrorOp) Run(_, _ io.Writer) error {
	return op.Err
}

func parseArgs(args []string) Op {
	provider := args[0]
	switch provider {
	case "-h", "--help":
		usage()
		os.Exit(0)
	case "aws", "amazon":
		return parseAWSArgs(args[1:])
	case "azr", "azure":
		return parseAzureArgs(args[1:])
	case "gcp", "google":
		return parseGCPArgs(args[1:])
	default:
		fmt.Printf("unrecognized cloud provider: %s\n", provider)
		os.Exit(1)
	}
	return nil
}

type Op interface {
	Run(stdout, stderr io.Writer) error
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("missing required cloud provider")
		usage()
		os.Exit(1)
	}

	op := parseArgs(os.Args[1:])

	if err := op.Run(os.Stdin, os.Stderr); err != nil {
		fmt.Println("unable to process command: ", err.Error())
		os.Exit(1)
	}
}

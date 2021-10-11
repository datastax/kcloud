package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func usage() {
	cmd := filepath.Base(os.Args[0])
	fmt.Printf("%s - retrieve kubernetes configuration from cloud providers\n", cmd)
	fmt.Printf("Usage: %s [options] aws|azr|gcp account [cluster-name]\n", cmd)
}

// ErrorOp indicates an unexpected error.
type ErrorOp struct{ Err error }

func (op ErrorOp) Run(_, _ io.Writer) error {
	return op.Err
}

// parseAWSArgs expects at least one argument which is the AWS profile to use.
// If additional args are provided, tries to interpret them as the region and cluster name.
func parseAWSArgs(args []string) Op {
	switch len(args) {
	case 0:
		fmt.Println("must specify aws profile name")
		return ErrorOp{fmt.Errorf("must specify aws profile name when using aws provider")}
	case 1:
		return awsListOp{
			profile: args[0],
		}
	case 2:
		regionCluster := strings.Split(args[1], awsClusterSep)
		if len(regionCluster) < 2 {
			fmt.Printf("invlid cluster specifier '%v', must be in the form region%vclusterName", args[1], awsClusterSep)
		}
		return awsUpdateOp{
			profile: args[0],
			region:  regionCluster[0],
			cluster: regionCluster[1],
		}
	default:
		return awsUpdateOp{
			profile: args[0],
			region:  args[1],
			cluster: args[2],
		}
	}
}

func parseArgs(args []string) Op {
	if len(args) < 2 {
		usage()
		os.Exit(1)
	}
	provider := args[0]
	switch provider {
	case "-h", "--help":
		usage()
		os.Exit(0)
	case "aws", "amazon":
		return parseAWSArgs(args[1:])
	case "azr", "azure":
		fmt.Println("azure is not yet implmented")
	case "gcp", "google":
		fmt.Println("gcp is not yet implemented")
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
	op := parseArgs(os.Args[1:])

	if err := op.Run(os.Stdin, os.Stderr); err != nil {
		fmt.Println("run returned an error")
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

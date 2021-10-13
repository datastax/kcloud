package main

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

const gcpCmd = "gcloud"

type gcpListOp struct {
	project string
}

type gcpUpdateOp struct {
	project string
	region  string
	cluster string
}

// parseGCPArgs expects at least one argument which is the GCP project to use.
func parseGCPArgs(args []string) Op {
	switch len(args) {
	case 0:
		return ErrorOp{fmt.Errorf("must specify gcp project name when using gcp cloud provider")}
	case 1:
		return gcpListOp{
			project: args[0],
		}
	default:
		region, cluster, err := parseQualifierCluster(args[1:])
		if err != nil {
			return ErrorOp{err}
		}
		return gcpUpdateOp{
			project: args[0],
			region:  region,
			cluster: cluster,
		}
	}
}

// Run lists the clusters available in the given project.
// runs the 'gcloud container clusters list' command
func (gcp gcpListOp) Run(stdout, stderr io.Writer) error {
	cmd := exec.Command(gcpCmd, "--project", gcp.project, "container", "clusters", "list")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("failed command: ", cmd)
		fmt.Print(string(output))
		return fmt.Errorf("failed to build aws command: %w", err)
	}
	clusters := parseGCPClusterList(output)
	for _, c := range clusters {
		fmt.Println(c)
	}
	return nil
}

// Run updates the kubeconfig file for the given region and cluster.
func (gcp gcpUpdateOp) Run(stdout, stderr io.Writer) error {
	cmd := exec.Command(gcpCmd, "--project", gcp.project, "container", "clusters", "get-credentials", gcp.cluster, "--region", gcp.region)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("failed command: ", cmd)
		fmt.Print(string(output))
		return fmt.Errorf("failed to build aws command: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func parseGCPClusterList(cmdOutput []byte) []string {
	clusters := []string{}
	for _, line := range strings.Split(string(cmdOutput), "\n") {
		if strings.HasPrefix(line, "NAME") {
			continue
		} else if strings.TrimSpace(line) == "" {
			break
		}
		parts := strings.Fields(line)
		clusters = append(clusters, parts[1]+clusterNameSep+parts[0])
	}
	return clusters
}

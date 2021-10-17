package main

import (
	"io"
)

const gcpCmd = "gcloud"

type gcpListProjectsOp struct{}

type gcpListClustersOp struct {
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
		return gcpListProjectsOp{}
	case 1:
		return gcpListClustersOp{
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

// Run lists the projects available to the current user
func (gcp gcpListProjectsOp) Run(stdout, stderr io.Writer) error {
	return RunCommand(gcpCmd, "projects", "list", "--format=value(projectId)")
}

// Run lists the clusters available in the given project.
// runs the 'gcloud container clusters list' command
func (gcp gcpListClustersOp) Run(stdout, stderr io.Writer) error {
	return RunCommand(gcpCmd, "--project", gcp.project, "container", "clusters", "list", "--format=value[separator=/](location,name)")
}

// Run updates the kubeconfig file for the given region and cluster.
func (gcp gcpUpdateOp) Run(stdout, stderr io.Writer) error {
	return RunCommand(gcpCmd, "--project", gcp.project, "container", "clusters", "get-credentials", gcp.cluster, "--region", gcp.region)
}

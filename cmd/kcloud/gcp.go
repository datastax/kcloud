package main

import (
	"github.com/alecthomas/kong"
)

const gcpCmd = "gcloud"

type GCPCmd struct {
	Project struct {
		Project string `arg:"" optional:""`
		Cluster struct {
			Cluster []string `arg:"" optional:"" name:"region/cluster"`
		} `arg:"" name:"region/cluster"`
	} `arg:""`
}

func (gcp *GCPCmd) Run(ctx *kong.Context) error {
	if gcp.Project.Project == "" {
		return RunCommandAndPrint(gcpCmd, "projects", "list", "--format=value(projectId)")
	}
	if len(gcp.Project.Cluster.Cluster) == 0 {
		const formatString = "--format=value[separator=" + clusterNameSep + "](location,name)"
		return RunCommandAndPrint(gcpCmd, "--project", gcp.Project.Project, "container", "clusters", "list", formatString)
	}
	region, cluster, err := parseQualifierCluster(gcp.Project.Cluster.Cluster)
	if err != nil {
		return err
	}
	return RunCommandAndPrint(gcpCmd, "--project", gcp.Project.Project, "container", "clusters", "get-credentials", "--region", region, cluster)
}

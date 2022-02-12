package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"

	"github.com/alecthomas/kong"
)

const azureCmd = "az"

type AzureCmd struct {
	Subscription struct {
		Subscription string `arg:"" optional:""`
		Cluster      struct {
			Cluster []string `arg:"" optional:"" name:"region/cluster"`
		} `arg:"" name:"region/cluster"`
	} `arg:""`
}

func (azr *AzureCmd) Run(ctx *kong.Context) error {
	if azr.Subscription.Subscription == "" {
		return RunCommand(azureCmd, "account", "list", "--query", "[].{id: id, name: name}", "--out", "tsv")
	}
	if len(azr.Subscription.Cluster.Cluster) == 0 {
		return AzureListClusters(ctx.Stdout, ctx.Stderr, azr.Subscription.Subscription)
	}
	resourceGroup, cluster, err := parseQualifierCluster(azr.Subscription.Cluster.Cluster)
	if err != nil {
		return err
	}
	return RunCommand(azureCmd, "aks", "get-credentials", "--overwrite-existing", "--subscription",
		azr.Subscription.Subscription, "--resource-group", resourceGroup, "--name", cluster)
}

type azureCluster struct {
	Name          string
	ResourceGroup string
}

// Run lists the clusters available in the given subscription.
// runs the 'az aks list' command
func AzureListClusters(stdout, stderr io.Writer, subscription string) error {
	args := []string{
		"aks", "list", "--subscription", subscription, "--query", "[].{name: name, resourceGroup: resourceGroup}",
	}
	cmd := exec.Command(azureCmd, args...)
	if cli.Verbose {
		PrintCommand(azureCmd, cmd.Args...)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("failed command: ", cmd)
		fmt.Print(string(output))
		return err
	}
	clusterList := []azureCluster{}
	if err := json.Unmarshal(output, &clusterList); err != nil {
		return err
	}
	for _, c := range clusterList {
		fmt.Println(c.ResourceGroup + clusterNameSep + c.Name)
	}
	return nil
}

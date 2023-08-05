package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/alecthomas/kong"
)

const azureCmd = "az"

type AzureCmd struct {
	Subscription struct {
		Subscription string `arg:"" optional:""`
		Cluster      struct {
			Cluster []string `arg:"" optional:"" name:"resource-group/cluster"`
		} `arg:"" name:"resource-group/cluster"`
	} `arg:""`
}

func (azr *AzureCmd) Run(ctx *kong.Context) error {
	if azr.Subscription.Subscription == "" {
		return RunCommandAndPrint(azureCmd, "account", "list", "--query", "[].{id: id, name: name}", "--out", "tsv")
	}
	if len(azr.Subscription.Cluster.Cluster) == 0 {
		return AzureListClusters(ctx.Stdout, ctx.Stderr, azr.Subscription.Subscription)
	}
	resourceGroup, cluster, err := parseQualifierCluster(azr.Subscription.Cluster.Cluster)
	if err != nil {
		return err
	}
	return RunCommandAndPrint(azureCmd, "aks", "get-credentials", "--overwrite-existing", "--subscription",
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
	output, err := RunCommand(azureCmd, args...)
	if err != nil {
		fmt.Print(string(output))
		fmt.Println("error: failed to run command: ", QuoteCommand(awsCmd, args...))
		return err
	}
	if cli.Verbose {
		fmt.Println("debug: raw command output")
		fmt.Print(string(output))
		fmt.Println("debug: end raw command output")
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

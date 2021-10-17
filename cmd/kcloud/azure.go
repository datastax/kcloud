package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

const azureCmd = "az"

type azureListSubscriptionsOp struct{}

type azureListClustersOp struct {
	subscription string
}

type azureUpdateConfigOp struct {
	subscription  string
	resourceGroup string
	cluster       string
}

// azureSubscriptionAliases provides a mapping from various names to the subscription UUID
// it's preferable to use UUID because it doesn't contain spaces or other problematic characters.
var azureSubscriptionAliases = map[string]string{
	// Astra Streaming Accounts
	"dev":                       "79f667ca-2aa1-4b8c-aa60-48e371d221c1",
	"Astra Streaming - Dev":     "79f667ca-2aa1-4b8c-aa60-48e371d221c1",
	"stage":                     "a3ea835b-d695-4224-adf9-3cabbd5d8953",
	"staging":                   "a3ea835b-d695-4224-adf9-3cabbd5d8953",
	"Astra Streaming - Staging": "a3ea835b-d695-4224-adf9-3cabbd5d8953",
	"prod":                      "6bb7fd0f-877d-4bcb-a2a8-fa6a4e8eab8d",
	"Astra Streaming - Prod":    "6bb7fd0f-877d-4bcb-a2a8-fa6a4e8eab8d",
	// CNDB
	"dev-cndb":               "bcc9b678-7b31-4af6-b705-53bbd21e0f46",
	"astra-serverless-dev-0": "bcc9b678-7b31-4af6-b705-53bbd21e0f46",
}

func lookupSubscriptionID(subscription string) string {
	if id, ok := azureSubscriptionAliases[subscription]; ok {
		return id
	}
	return subscription
}

// parseAzureArgs expects at least one argument which is the Azure project to use.
func parseAzureArgs(args []string) Op {
	switch len(args) {
	case 0:
		return azureListSubscriptionsOp{}
	case 1:
		return azureListClustersOp{
			subscription: args[0],
		}
	default:
		resourceGroup, cluster, err := parseQualifierCluster(args[1:])
		if err != nil {
			return ErrorOp{err}
		}
		return azureUpdateConfigOp{
			subscription:  args[0],
			resourceGroup: resourceGroup,
			cluster:       cluster,
		}
	}
}

type AzureSubscription struct {
	Name string
	Id   string
}

// Run lists the subscriptions available to the user.
// runs the 'gcloud container clusters list' command
func (azr azureListSubscriptionsOp) Run(stdout, stderr io.Writer) error {
	return RunCommand(azureCmd, "account", "list", "--query", "[].{id: id, name: name}", "--out", "tsv")
}

// Run lists the clusters available in the given subscription.
// runs the 'az aks list' command
func (azr azureListClustersOp) Run(stdout, stderr io.Writer) error {
	cmd := exec.Command(azureCmd, "aks", "list", "--subscription", lookupSubscriptionID(azr.subscription), "--query", "[].{name: name, resourceGroup: resourceGroup}")
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

// Run updates the kubeconfig file for the given region and cluster.
func (azr azureUpdateConfigOp) Run(stdout, stderr io.Writer) error {
	return RunCommand(azureCmd, "aks", "get-credentials", "--overwrite-existing", "--subscription", lookupSubscriptionID(azr.subscription),
		"--resource-group", azr.resourceGroup, "--name", azr.cluster)
}

type azureCluster struct {
	Name          string
	ResourceGroup string
}

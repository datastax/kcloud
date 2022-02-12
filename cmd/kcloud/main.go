package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
)

// cli represents the parsed command line
var cli struct {
	Verbose bool `help:"verbose mode." default:"false" short:"v"`

	Gcp GCPCmd `cmd:"" help:"GCP provider" aliases:"google,gke"`

	Azr AzureCmd `cmd:"" help:"Azure provider" aliases:"azure,aks"`

	Aws AWSCmd `cmd:"" help:"AWS provider" aliases:"amazon,aks"`
}

const description = "Download kubeconfig for various kubernetes cloud providers.\nExample: kcloud gcp my-dev-project us-east4/my-awesome-cluster"

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("kcloud"),
		kong.Description(description),
		kong.UsageOnError())
	err := ctx.Run()
	if err != nil {
		fmt.Println("command failed: ", err.Error())
		os.Exit(1)
	}
}

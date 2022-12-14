# kcloud

This command provides a simplified CLI to download kubernetes config from various
cloud providers.

## Prerequisites

A local installation of the cloud provider commands `aws`, `az`, and/or `gcloud` is required.  Before using
`kcloud`, check that you can successfully run the relevant provider specific commands.

## Install

### Option 1 - Use the `go install` command

To install using this method, you must have the go compiler installed, and have `~/go/bin` in your PATH.

    go install github.com/datastax/kcloud/cmd/kcloud@latest

### Option 2 - Download the binary

The latest release binaries can be downloaded from the github releases page.

    https://github.com/datastax/kcloud/releases

## Usage

    kcloud aws|azr|gcp [profile|subscription|project] [region/cluster]
    
### AWS Example

    kcloud aws dev us-east-2/my-cluster
    
### Azure Example

    kcloud azr 79f667ca-2aa1-4b8c-aa60-48e371d321c1 dev-resource-group/my-cluster
    
### GCP Example

    kcloud gcp my-project-dev us-east4/my-cluster
    
## Build From Source

    make install

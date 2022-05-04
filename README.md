# kcloud

CLI tool to download kube config settings from various cloud providers.  Requires local installation
of aws, az, and/or gcloud commands to access the cloud providers.

## Install

    go install github.com/riptano/kcloud/cmd/kcloud@latest

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

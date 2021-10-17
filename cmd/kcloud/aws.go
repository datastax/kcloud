package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type awsListProfilesOp struct{}

type awsListClustersOp struct {
	profile string
}

type awsUpdateOp struct {
	profile string
	region  string
	cluster string
}

var awsRegions = map[string]bool{
	"us-east-1": true,
	"us-east-2": true,
	"us-west-2": true,
}

const awsCmd = "aws"

// parseAWSArgs expects at least one argument which is the AWS profile to use.
// If additional args are provided, tries to interpret them as the region and cluster name.
func parseAWSArgs(args []string) Op {
	switch len(args) {
	case 0:
		return awsListProfilesOp{}
	case 1:
		return awsListClustersOp{
			profile: args[0],
		}
	default:
		region, cluster, err := parseQualifierCluster(args[1:])
		if err != nil {
			return ErrorOp{err}
		}
		return awsUpdateOp{
			profile: args[0],
			region:  region,
			cluster: cluster,
		}
	}
}

// Run lists the available profiles
func (aws awsListProfilesOp) Run(stdout, stderr io.Writer) error {
	profiles, err := awsListAvailableProfiles(DefaultAWSCredsFilePath())
	if err != nil {
		return fmt.Errorf("unable to parse AWS credentials: %w", err)
	}
	for _, profile := range profiles {
		fmt.Println(profile)
	}
	return nil
}

// Run lists the clusters available to the given profile in the known regions.
// runs the 'aws eks list-clusters' command once for each region in parallel
func (aws awsListClustersOp) Run(stdout, stderr io.Writer) error {
	clusters := []string{}
	wg := sync.WaitGroup{}
	var groupErr error
	for region := range awsRegions {
		wg.Add(1)
		go func(profile, region string) {
			defer wg.Done()
			regionClusters, err := awsListClustersInRegion(profile, region)
			if err != nil {
				groupErr = err
				return
			}
			for _, c := range regionClusters {
				clusters = append(clusters, region+clusterNameSep+c)
			}
		}(aws.profile, region)
	}
	wg.Wait()
	if groupErr != nil {
		return groupErr
	}
	for _, c := range clusters {
		fmt.Println(c)
	}
	return nil
}

// Run updates the kubeconfig file for the given region and cluster.
func (aws awsUpdateOp) Run(stdout, stderr io.Writer) error {
	if !awsRegions[aws.region] {
		fmt.Printf("warning: unrecognized region %v\n", aws.region)
	}
	return RunCommand(awsCmd, "--profile", aws.profile, "eks", "--region", aws.region, "update-kubeconfig", "--name", aws.cluster)
}

type awsCreds struct {
	profile         string
	accessKeyID     string
	secretAccessKey string
}

func awsListClustersInRegion(profile string, region string) ([]string, error) {
	cmd := exec.Command(awsCmd, "--profile", profile, "eks", "--region", region, "list-clusters")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("failed command: ", cmd)
		fmt.Print(string(output))
		return nil, fmt.Errorf("failed to build aws command: %w", err)
	}

	return awsParseClusterList(output)
}

type awsClusterList struct {
	Clusters []string `json:"clusters"`
}

func awsParseClusterList(awsCmdOut []byte) ([]string, error) {
	clusterList := awsClusterList{}
	if err := json.Unmarshal(awsCmdOut, &clusterList); err != nil {
		return nil, err
	}
	return clusterList.Clusters, nil
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alecthomas/kong"
)

type AWSCmd struct {
	Profile struct {
		Profile string `arg:"" optional:""`
		Cluster struct {
			Cluster []string `arg:"" optional:"" name:"region/cluster"`
		} `arg:"" name:"region/cluster"`
	} `arg:""`
}

func (aws *AWSCmd) Run(ctx *kong.Context) error {
	if aws.Profile.Profile == "" {
		return AWSListProfiles()
	}
	if len(aws.Profile.Cluster.Cluster) == 0 {
		return aws.AWSListClusters()
	}
	region, cluster, err := parseQualifierCluster(aws.Profile.Cluster.Cluster)
	if err != nil {
		return err
	}
	if !awsRegions[region] {
		fmt.Printf("WARNING: unrecognized region %v\n", region)
	}
	return RunCommandAndPrint(awsCmd, "--profile", aws.Profile.Profile, "eks", "--region", region,
		"update-kubeconfig", "--name", cluster)

}

var awsRegions = map[string]bool{
	"us-east-1": true,
	"us-east-2": true,
	"us-west-2": true,
}

const awsCmd = "aws"

// AWSListProfiles lists the available profiles
func AWSListProfiles() error {
	profiles, err := awsListAvailableProfiles(DefaultAWSConfigFilePath())
	if err != nil {
		return fmt.Errorf("unable to parse AWS credentials: %w", err)
	}
	for _, profile := range profiles {
		fmt.Println(profile)
	}
	return nil
}

// AWSListClusters lists the clusters available to the given profile in the known regions.
// runs the 'aws eks list-clusters' command once for each region in parallel
func (aws *AWSCmd) AWSListClusters() error {
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
		}(aws.Profile.Profile, region)
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

func awsListClustersInRegion(profile string, region string) ([]string, error) {
	output, err := RunCommand(awsCmd, "--profile", profile, "eks", "--region", region, "list-clusters")
	if err != nil {
		fmt.Print(string(output))
		fmt.Println("ERROR: failed command: ", QuoteCommand(awsCmd, "--profile", profile, "eks", "--region", region, "list-clusters"))
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

func DefaultAWSConfigFilePath() string {
	const awsDefaultConfigPath = ".aws/config"
	homedir, _ := os.UserHomeDir()
	return filepath.Join(homedir, awsDefaultConfigPath)
}

func awsListAvailableProfiles(configFile string) ([]string, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open AWS creds file '%s': %w", configFile, err)
	}
	profiles := []string{}
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "[") {
			profile := strings.TrimRight(strings.TrimLeft(line, "["), "]")
			profiles = append(profiles, profile)
		}
	}
	return profiles, nil
}

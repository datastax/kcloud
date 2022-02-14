package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
		return AWSPrintConfigProfiles()
	}
	if len(aws.Profile.Cluster.Cluster) == 0 {
		return aws.AWSListClusters()
	}
	region, cluster, err := parseQualifierCluster(aws.Profile.Cluster.Cluster)
	if err != nil {
		return err
	}
	if _, ok := awsKnownRegions[region]; !ok {
		fmt.Printf("WARNING: unrecognized region %v\n", region)
	}
	return RunCommandAndPrint(awsCmd, "--profile", aws.Profile.Profile, "eks", "--region", region,
		"update-kubeconfig", "--name", cluster)

}

// awsKnownRegions as defined in the AWS docs
// https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html
var awsKnownRegions = map[string]struct{}{
	"us-east-2":      {},
	"us-east-1":      {},
	"us-west-1":      {},
	"us-west-2":      {},
	"af-south-1":     {},
	"ap-east-1":      {},
	"ap-southeast-3": {},
	"ap-south-1":     {},
	"ap-northeast-3": {},
	"ap-northeast-2": {},
	"ap-southeast-1": {},
	"ap-southeast-2": {},
	"ap-northeast-1": {},
	"ca-central-1":   {},
	"eu-central-1":   {},
	"eu-west-1":      {},
	"eu-west-2":      {},
	"eu-south-1":     {},
	"eu-west-3":      {},
	"eu-north-1":     {},
	"me-south-1":     {},
	"sa-east-1":      {},
	"us-gov-east-1":  {},
	"us-gov-west-1":  {},
}

const awsCmd = "aws"

// AWSPrintConfigProfiles lists the available profiles
func AWSPrintConfigProfiles() error {
	config, err := awsLoadConfig(DefaultAWSConfigFilePath())
	if err != nil {
		return fmt.Errorf("unable to parse AWS credentials: %w", err)
	}
	for _, profile := range config.profiles {
		fmt.Println(profile)
	}
	return nil
}

// AWSListClusters lists the clusters available to the given profile in the known regions.
// runs the 'aws eks list-clusters' command once for each region in parallel
func (aws *AWSCmd) AWSListClusters() error {
	awsConfig, err := awsLoadConfig(DefaultAWSConfigFilePath())
	if err != nil {
		fmt.Println("ERROR: unable to load AWS config file: " + err.Error())
	}
	clusters := []string{}
	wg := sync.WaitGroup{}
	var groupErr error
	for region := range awsConfig.regions {
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

// awsClusterList is used for unmarshalling the aws command output
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

// awsProfileRegex matches a line like "[myprofile]"
var awsProfileRegex = regexp.MustCompile(`\[([^\]]+)\]`)

// awsRegionRegex matches a line like "region = useast1"
var awsRegionRegex = regexp.MustCompile(`region\s*=\s*(.+)`)

type awsConfig struct {
	profiles []string
	regions  map[string]struct{}
}

func awsLoadConfig(configFile string) (awsConfig, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return awsConfig{}, fmt.Errorf("unable to open AWS creds file '%s': %w", configFile, err)
	}
	defer f.Close()

	awsConfig := awsConfig{
		profiles: []string{},
		regions:  map[string]struct{}{},
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if match := awsProfileRegex.FindStringSubmatch(line); len(match) > 1 {
			profile := strings.TrimSpace(match[1])
			awsConfig.profiles = append(awsConfig.profiles, profile)
		}
		if match := awsRegionRegex.FindStringSubmatch(line); len(match) > 1 {
			region := strings.TrimSpace(match[1])
			awsConfig.regions[region] = struct{}{}
		}
	}
	return awsConfig, nil
}

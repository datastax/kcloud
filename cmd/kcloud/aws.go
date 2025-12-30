package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
	AllRegions bool   `short:"a" help:"Search all knows AWS regions instead of just regions found in ~/.aws/config"`
	RoleARN    string `short:"r" help:"Include a role arn when updating kube config"`
}

func (aws *AWSCmd) Run(ctx *kong.Context) error {
	if aws.Profile.Profile == "" {
		return AWSPrintConfigProfiles(DefaultAWSConfigFilePath())
	}
	if len(aws.Profile.Cluster.Cluster) == 0 {
		return aws.AWSListClusters()
	}
	region, cluster, err := parseQualifierCluster(aws.Profile.Cluster.Cluster)
	if err != nil {
		return err
	}
	if _, ok := awsKnownRegions[region]; !ok {
		fmt.Printf("warning: unrecognized region argument '%s'\n", region)
	}
	args := []string{"--profile", aws.Profile.Profile, "eks", "--region", region,
		"update-kubeconfig", "--name", cluster}
	if strings.TrimSpace(aws.RoleARN) != "" {
		args = append(args, "--role-arn", strings.TrimSpace(aws.RoleARN))
	}
	return RunCommandAndPrint(awsCmd, args...)

}

// awsKnownRegions as defined in the AWS docs
// https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html
// Last updated: December 2024
var awsKnownRegions = map[string]struct{}{
	// US Regions
	"us-east-2": {}, // Ohio
	"us-east-1": {}, // N. Virginia
	"us-west-1": {}, // N. California
	"us-west-2": {}, // Oregon

	// Africa
	"af-south-1": {}, // Cape Town

	// Asia Pacific
	"ap-east-1":      {}, // Hong Kong
	"ap-south-1":     {}, // Mumbai
	"ap-south-2":     {}, // Hyderabad
	"ap-southeast-1": {}, // Singapore
	"ap-southeast-2": {}, // Sydney
	"ap-southeast-3": {}, // Jakarta
	"ap-southeast-4": {}, // Melbourne
	"ap-southeast-5": {}, // Malaysia
	"ap-northeast-1": {}, // Tokyo
	"ap-northeast-2": {}, // Seoul
	"ap-northeast-3": {}, // Osaka

	// Canada
	"ca-central-1": {}, // Central
	"ca-west-1":    {}, // Calgary

	// Europe
	"eu-central-1": {}, // Frankfurt
	"eu-central-2": {}, // Zurich
	"eu-west-1":    {}, // Ireland
	"eu-west-2":    {}, // London
	"eu-west-3":    {}, // Paris
	"eu-south-1":   {}, // Milan
	"eu-south-2":   {}, // Spain
	"eu-north-1":   {}, // Stockholm

	// Middle East
	"il-central-1": {}, // Tel Aviv
	"me-south-1":   {}, // Bahrain
	"me-central-1": {}, // UAE

	// South America
	"sa-east-1": {}, // São Paulo

	// AWS GovCloud (US)
	"us-gov-east-1": {}, // US-East
	"us-gov-west-1": {}, // US-West
}

const awsCmd = "aws"

// AWSPrintConfigProfiles lists the available profiles in the given config file
func AWSPrintConfigProfiles(configFile string) error {
	config, err := awsLoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("unable to parse AWS config file: %w", err)
	}
	for _, profile := range config.profiles {
		// Strip "profile " prefix if present for consistency with AWS CLI
		profile = strings.TrimPrefix(profile, "profile ")
		fmt.Println(profile)
	}
	return nil
}

// AWSListClusters lists the clusters available to the given profile in the known regions.
// runs the 'aws eks list-clusters' command once for each region in parallel
func (aws *AWSCmd) AWSListClusters() error {
	regions := awsKnownRegions
	if !aws.AllRegions {
		awsConfig, err := awsLoadConfig(DefaultAWSConfigFilePath())
		if err != nil {
			return fmt.Errorf("error: unable to load AWS config file: %w", err)
		}
		regions = awsConfig.regions
	}
	var mu sync.Mutex
	clusters := []string{}
	wg := sync.WaitGroup{}
	for region := range regions {
		wg.Add(1)
		go func(profile, region string) {
			defer wg.Done()
			regionClusters, err := awsListClustersInRegion(profile, region)
			if err != nil {
				fmt.Printf("error: failed to search region '%s':\n %s\n", region, err.Error())
				return
			}
			// Prefix each cluster name with region
			qualifiedClusters := make([]string, len(regionClusters))
			for i, c := range regionClusters {
				qualifiedClusters[i] = region + clusterNameSep + c
			}
			mu.Lock()
			clusters = append(clusters, qualifiedClusters...)
			mu.Unlock()
		}(aws.Profile.Profile, region)
	}
	wg.Wait()
	sort.Strings(clusters)
	for _, c := range clusters {
		fmt.Println(c)
	}
	return nil
}

func awsListClustersInRegion(profile string, region string) ([]string, error) {
	args := []string{"--profile", profile, "eks", "--region", region, "list-clusters"}
	output, err := RunCommand(awsCmd, args...)
	if err != nil {
		return nil, fmt.Errorf("failed command: %s\n  output: %s\n  err: %w",
			QuoteCommand(awsCmd, args...),
			strings.TrimSpace(string(output)),
			err,
		)
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
var awsRegionRegex = regexp.MustCompile(`region\s*=\s*([\w-]+)`)

type awsConfig struct {
	profiles []string
	regions  map[string]struct{}
}

func awsLoadConfig(configFile string) (*awsConfig, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open AWS creds file '%s': %w", configFile, err)
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
			// Normalize profile name by stripping "profile " prefix
			// AWS config uses [default] and [profile name] formats
			normalizedProfile := strings.TrimPrefix(profile, "profile ")
			if stringInSlice(normalizedProfile, awsConfig.profiles) {
				return nil, fmt.Errorf("invalid aws config, found duplicate profile '%v'", normalizedProfile)
			}
			awsConfig.profiles = append(awsConfig.profiles, normalizedProfile)
		}
		if match := awsRegionRegex.FindStringSubmatch(line); len(match) > 1 {
			region := strings.TrimSpace(match[1])
			if _, ok := awsKnownRegions[region]; !ok {
				fmt.Printf("warning: unrecognized AWS region: '%s'\n", region)
			}
			awsConfig.regions[region] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	return &awsConfig, nil
}

func stringInSlice(s string, xs []string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

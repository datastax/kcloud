package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type awsListOp struct {
	profile string
}

type awsUpdateOp struct {
	profile string
	region  string
	cluster string
}

var awsRegions = []string{
	"us-east-1",
	"us-east-2",
	"us-west-2",
}

// awsClusterSep is a separator used for the aws region and cluster name
const awsClusterSep = "/"

// Run lists the clusters available to the given profile in the known regions.
// runs the 'aws eks list-clusters' command once for each region in parallel
func (aws awsListOp) Run(stdout, stderr io.Writer) error {
	clusters := []string{}
	wg := sync.WaitGroup{}
	var groupErr error
	for _, region := range awsRegions {
		// TODO: could these calls be made in parallel?
		wg.Add(1)
		go func(profile, region string) {
			defer wg.Done()
			regionClusters, err := awsListClustersByRegion(profile, region)
			if err != nil {
				groupErr = err
				return
			}
			for _, c := range regionClusters {
				clusters = append(clusters, region+awsClusterSep+c)
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
	cmd := exec.Command("aws", "--profile", aws.profile, "eks", "--region", aws.region, "update-kubeconfig", "--name", aws.cluster)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("failed command: ", cmd)
		fmt.Print(string(output))
		return fmt.Errorf("failed to build aws command: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

type awsCreds struct {
	profile         string
	accessKeyID     string
	secretAccessKey string
}

// TODO: resolve user home directory
const defaultAWSCredsFile = "/Users/paulgier/.aws/credentials"

func awsReadCredentials(credsFile string) ([]awsCreds, error) {
	f, err := os.Open(credsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open AWS creds file %v: %w", credsFile, err)
	}
	creds := []awsCreds{}
	scanner := bufio.NewScanner(f)

	var section []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") {
			continue
		} else if strings.HasPrefix(line, "[") {
			if len(section) > 0 {
				creds = append(creds, awsParseCredentials(section))
			}
			section = []string{line}
		} else if line != "" {
			section = append(section, line)
		}
	}
	creds = append(creds, awsParseCredentials(section))
	return creds, nil
}

func awsParseCredentials(section []string) awsCreds {
	creds := awsCreds{
		profile: strings.TrimSuffix(strings.TrimPrefix(section[0], "["), "]"),
	}
	for _, line := range section[1:] {
		pair := strings.Split(line, "=")
		if len(pair) > 1 {
			switch strings.TrimSpace(pair[0]) {
			case "aws_access_key_id":
				creds.accessKeyID = strings.TrimSpace(pair[1])
			case "aws_secret_access_key":
				creds.secretAccessKey = strings.TrimSpace(pair[1])
			}
		}

	}
	return creds
}

func awsListClustersByRegion(profile string, region string) ([]string, error) {
	cmd := exec.Command("aws", "--profile", profile, "eks", "--region", region, "list-clusters")
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

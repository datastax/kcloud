package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultAWSCredsFile = ".aws/credentials"

func DefaultAWSCredsFilePath() string {
	homedir, _ := os.UserHomeDir()
	return filepath.Join(homedir, defaultAWSCredsFile)
}

func awsListAvailableProfiles(credsFile string) ([]string, error) {
	f, err := os.Open(credsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to open AWS creds file %v: %w", credsFile, err)
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

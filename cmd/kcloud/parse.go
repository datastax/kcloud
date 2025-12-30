package main

import (
	"fmt"
	"strings"
)

const clusterNameSep = "/"

// parseQualifierCluster tries to find qualifier and cluster which
// can either be two separate strings, or a single string separated by a slash
func parseQualifierCluster(args []string) (string, string, error) {
	switch len(args) {
	case 0:
		return "", "", fmt.Errorf("invalid cluster specifier, requires at least 1 arg, received 0")
	case 1:
		parts := strings.SplitN(args[0], clusterNameSep, 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid cluster specifier '%v', must be in the form region%vclusterName", args[0], clusterNameSep)
		}
		return parts[0], parts[1], nil
	default:
		return args[0], args[1], nil
	}
}

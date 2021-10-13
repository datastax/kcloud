package main

import (
	"fmt"
	"strings"
)

const clusterNameSep = "/"

func parseQualifierCluster(args []string) (string, string, error) {
	switch len(args) {
	case 0:
		return "", "", fmt.Errorf("requires at least 1 cluster arg, received 0")
	case 1:
		parts := strings.SplitN(args[0], clusterNameSep, 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invlid cluster specifier '%v', must be in the form region%vclusterName", args[0], clusterNameSep)
		}
		return parts[0], parts[1], nil
	default:
		return args[0], args[1], nil
	}
}

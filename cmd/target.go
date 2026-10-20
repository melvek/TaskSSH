package cmd

import (
	"fmt"
	"strings"
)

func setTargetHostPattern(args []string) error {
	targetHostPattern = ""
	if len(args) != 1 {
		return fmt.Errorf("specify exactly one host pattern as a positional argument")
	}
	if strings.TrimSpace(args[0]) == "" {
		return fmt.Errorf("target hosts cannot be empty")
	}
	targetHostPattern = args[0]
	return nil
}

package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func confirm(message string) bool {
	fmt.Printf("%s (y/n): ", message)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

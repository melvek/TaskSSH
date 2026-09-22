package main

import (
	"fmt"
	"os"

	"mestrap.com/taskssh/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

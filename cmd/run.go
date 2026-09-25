package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"mestrap.com/taskssh/internal/action"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/log"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/ssh"
	"mestrap.com/taskssh/internal/task"
)

// runBuiltinTask 执行内置任务。
func runBuiltinTask(taskName string, hosts []string, with map[string]any) error {
	inv, err := config.Load(inventoryFile)
	if err != nil {
		return err
	}

	step := config.Step{
		Name:   taskName,
		Action: taskName,
		With:   with,
	}
	t := config.Task{
		Steps: []config.Step{step},
	}

	return executeTask(&t, hosts, inv)
}

// executeTask 是公共执行入口。
func executeTask(t *config.Task, hostNames []string, inv *config.Inventory) error {
	ssh.ClearPasswordCache()
	defer ssh.ClearPasswordCache()

	ov := task.Overrides{
		Port:     portFlag,
		Username: userFlag,
		Password: passwordFlag,
	}

	entries, err := task.ResolveHosts(hostNames, inv, ov)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("no target hosts")
	}

	log.Section("Target hosts")
	for _, e := range entries {
		log.ListItem(e.Name, e.Host.Host)
	}

	if listOnly && !dryRun {
		return nil
	}

	if !yesFlag && !dryRun {
		if !confirm("Confirm to proceed") {
			return nil
		}
	}

	globalVars := make(resolve.Vars)
	for k, v := range inv.GlobalVars.Extra {
		globalVars[k] = v
	}

	actions := action.NewRegistry()
	executor := task.NewExecutor(actions, concurrency, connectTimeout, dryRun)
	results := executor.Run(t, entries, globalVars)

	if dryRun {
		return nil
	}

	success, failed := 0, 0
	var failedHosts []string
	for _, r := range results {
		if r.Success {
			success++
		} else {
			failed++
			failedHosts = append(failedHosts, r.Host)
		}
	}

	log.Summary(len(results), success, failed)
	if len(failedHosts) > 0 {
		log.Hint("Failed hosts: %v", failedHosts)
	}

	if failed > 0 {
		return fmt.Errorf("%d host(s) failed", failed)
	}
	return nil
}

// confirm 询问用户。
func confirm(message string) bool {
	fmt.Printf("\n%s (y/n): ", message)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

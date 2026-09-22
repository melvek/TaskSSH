package cmd

import (
	"fmt"
	"os"

	"mestrap.com/taskssh/internal/action"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/log"
	"mestrap.com/taskssh/internal/resolve"
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
	// 构造 CLI 覆盖
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

	// 展示主机列表
	log.Section("Target hosts")
	for _, e := range entries {
		log.ListItem(e.Name, fmt.Sprintf("%s@%s", e.Host.Username, e.Host.Host))
	}
	log.EmptyLine()

	if listOnly {
		return nil
	}

	if !yesFlag {
		if !confirm("Confirm to proceed") {
			return nil
		}
	}

	globalVars := make(resolve.Vars)
	for k, v := range inv.GlobalVars.Extra {
		globalVars[k] = v
	}

	actions := action.NewRegistry()
	executor := task.NewExecutor(actions)
	results := executor.Run(t, entries, globalVars)

	success, failed := 0, 0
	for _, r := range results {
		if r.Success {
			success++
		} else {
			failed++
		}
	}
	log.Section("Summary")
	log.Info("Total: %d, Success: %d, Failed: %d", len(results), success, failed)

	if failed > 0 {
		os.Exit(1)
	}
	return nil
}

// confirm 询问用户。
func confirm(message string) bool {
	fmt.Printf("%s (y/n): ", message)
	var input string
	fmt.Scanln(&input)
	return input == "y" || input == "yes"
}

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"maps"
	"os"
	"strings"

	"mestrap.com/taskssh/internal/action"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/ssh"
	"mestrap.com/taskssh/internal/task"
)

// reservedVars 是运行时会注入的变量，CLI 不允许覆盖。
var reservedVars = map[string]bool{
	"date":   true,
	"execId": true,
}

// runBuiltinTask 执行内置任务
func runBuiltinTask(ctx context.Context, taskName string, hosts []string, with any) error {
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

	return executeTask(ctx, &t, hosts, inv)
}

// executeTask 是公共执行入口。
func executeTask(ctx context.Context, t *config.Task, hostNames []string, inv *config.Inventory) error {
	ssh.ClearPasswordCache()
	defer ssh.ClearPasswordCache()

	// dry-run 提醒
	if dryRun {
		console.Warn("Dry run (no changes will be made)")
	}

	cliVarMap, err := parseCLIVars(cliVars)
	if err != nil {
		return err
	}

	ov := task.Overrides{
		Port:         portFlag,
		Username:     userFlag,
		Password:     passwordFlag,
		IdentityFile: identityFileFlag,
	}

	entries, err := task.ResolveHosts(hostNames, inv, ov)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("no target hosts")
	}

	lines := make([]string, 0, len(entries))
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%s: %s", e.Name, e.Host.Host))
	}
	console.Targets(lines)

	if listOnly {
		return nil
	}

	// dry run 不需要确认操作
	if !yesFlag && !dryRun {
		if !confirm("Confirm to proceed") {
			return nil
		}
	}

	globalVars := make(resolve.Vars)
	maps.Copy(globalVars, inv.GlobalVars.Extra)

	// 一次性加载所有Action，之后根据任务创建执行器
	actions := action.NewRegistry()
	serial := t.Serial
	if serialFlag >= 0 {
		serial = serialFlag
	}

	minAvailable := t.MinAvailable
	if minAvailableFlag >= 0 {
		minAvailable = minAvailableFlag
	}

	executor := task.NewExecutor(actions, concurrency, connectTimeout, dryRun, cliVarMap,
		serial, minAvailable, stopOnFailureFlag)

	results, err := executor.Run(ctx, t, entries, globalVars)
	if err != nil {
		return err
	}

	if dryRun {
		return nil
	}

	// 整理执行结果，输出执行摘要信息
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

	console.Summary(len(results), success, failed)
	if len(failedHosts) > 0 {
		console.Hint("Failed hosts: %v", failedHosts)
	}

	console.Info("----------------------")
	console.Info("Execution ID: %s", executor.ExecID())
	console.EmptyLine()

	if failed > 0 {
		return fmt.Errorf("%d host(s) failed", failed)
	}

	return nil
}

// parseCLIVars 解析 -D 传入的 key=value 变量。
//
// 规则：
//   - 格式必须是 key=value，value 可含 '='
//   - key 和 value 均做 trim
//   - key 不能为空
//   - 不能覆盖保留变量
//   - 重复的 key 以后出现的为准
func parseCLIVars(vars []string) (resolve.Vars, error) {
	out := make(resolve.Vars, len(vars))

	for _, kv := range vars {
		before, after, ok := strings.Cut(kv, "=")
		if !ok {
			return nil, fmt.Errorf("invalid -D %q, expected key=value", kv)
		}

		key := strings.TrimSpace(before)
		val := strings.TrimSpace(after)

		if key == "" {
			return nil, fmt.Errorf("invalid -D %q, key is empty", kv)
		}

		if reservedVars[key] {
			return nil, fmt.Errorf(
				"%s is a reserved variable and cannot be overridden by -D", key)
		}

		out[key] = val
	}

	return out, nil
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

package task

import (
	"bytes"
	"fmt"
	"maps"
	"strings"
	"sync"
	"time"

	"mestrap.com/taskssh/internal/action"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/resolve"
	"mestrap.com/taskssh/internal/ssh"
)

// Result 是单台主机的执行结果。
type Result struct {
	Host    string
	Success bool
	Error   error
}

// Executor 执行任务。
type Executor struct {
	actions        *action.Registry
	concurrency    int
	connectTimeout time.Duration
	dryRun         bool
	cliVars        resolve.Vars
	execID         string
	serial         int
	minAvailable   int
	stopOnFailure  bool
}

// Batch 是一批待执行的主机，属于同一组。
type Batch struct {
	Group string
	Hosts []HostEntry
}

// NewExecutor 创建执行器。
//
// cliVars 是 CLI -D 传入的变量，优先级最高，覆盖清单中的所有同名变量。
func NewExecutor(actions *action.Registry, concurrency int, connectTimeout int,
	dryRun bool, cliVars resolve.Vars,
	serial, minAvailable int, stopOnFailure bool) *Executor {
	if concurrency <= 0 {
		concurrency = 1
	}
	if connectTimeout <= 0 {
		connectTimeout = 10
	}
	if cliVars == nil {
		cliVars = make(resolve.Vars)
	}
	return &Executor{
		actions:        actions,
		concurrency:    concurrency,
		connectTimeout: time.Duration(connectTimeout) * time.Second,
		dryRun:         dryRun,
		cliVars:        cliVars,
		serial:         serial,
		minAvailable:   minAvailable,
		stopOnFailure:  stopOnFailure,
	}
}

// ExecID 返回本次执行的唯一标识。
//
// 仅在 Run 执行后有效。
func (e *Executor) ExecID() string {
	return e.execID
}

// groupEntries 按 Name 的 "/" 前缀分组，返回组顺序和分组结果。
func groupEntries(entries []HostEntry) ([]string, map[string][]HostEntry) {
	order := []string{}
	groups := map[string][]HostEntry{}

	for _, e := range entries {
		g := "standalone"
		if i := strings.Index(e.Name, "/"); i > 0 {
			g = e.Name[:i]
		}
		if _, ok := groups[g]; !ok {
			order = append(order, g)
		}
		groups[g] = append(groups[g], e)
	}
	return order, groups
}

// planBatches 把主机列表规划成一条批队列。
//
// 规则：
//   - 按组分块，组间顺序按 entries 里组第一次出现的顺序
//   - 组内按 batchSize = min(serial, len(group)-minAvailable) 切片
//   - 单台组单独一批，不参与 minAvailable 计算
//   - len(group)-minAvailable <= 0 时报错
func planBatches(entries []HostEntry, serial, minAvailable int) ([]Batch, error) {
	if serial < 0 {
		serial = 0
	}
	if minAvailable < 0 {
		minAvailable = 0
	}

	order, groups := groupEntries(entries)
	var plan []Batch

	for _, gname := range order {
		groupHosts := groups[gname]

		// 单台组：直接一批
		if len(groupHosts) == 1 {
			plan = append(plan, Batch{Group: gname, Hosts: groupHosts})
			continue
		}

		available := len(groupHosts) - minAvailable
		if available <= 0 {
			return nil, fmt.Errorf(
				"group %s: min-available %d >= group size %d, cannot batch\n"+
					"hint: reduce min-available in task config or override with --min-available",
				gname, minAvailable, len(groupHosts))
		}

		batchSize := available
		if serial > 0 {
			batchSize = min(serial, available)
		}

		for start := 0; start < len(groupHosts); start += batchSize {
			end := start + batchSize
			if end > len(groupHosts) {
				end = len(groupHosts)
			}
			plan = append(plan, Batch{Group: gname, Hosts: groupHosts[start:end]})
		}
	}

	return plan, nil
}

// Run 对一批主机执行任务。
func (e *Executor) Run(task *config.Task, hosts []HostEntry, globalVars resolve.Vars) ([]Result, error) {
	if !e.dryRun {
		console.Section("Start")
	}

	if task.Description != "" {
		console.Info("Task: %s (%d steps)", task.Description, len(task.Steps))
	} else {
		console.Info("Steps: %d", len(task.Steps))
	}

	console.Info("Targets: %d", len(hosts))
	if !e.dryRun {
		console.Info("Concurrency: %d", e.concurrency)
	}

	e.execID = generateExecID()

	// 启用分批：serial > 0 或 minAvailable > 0
	if e.serial > 0 || e.minAvailable > 0 {
		plan, err := planBatches(hosts, e.serial, e.minAvailable)
		if err != nil {
			return nil, err
		}
		if e.dryRun {
			printPlan(plan)
		}
		return e.runBatches(task, plan, globalVars), nil
	}

	if e.concurrency == 1 || e.dryRun {
		return e.runSerial(task, hosts, globalVars), nil
	}
	return e.runParallel(task, hosts, globalVars), nil
}

// runSerial 串行执行，实时输出。
func (e *Executor) runSerial(task *config.Task, hosts []HostEntry, globalVars resolve.Vars) []Result {
	results := make([]Result, 0, len(hosts))

	for i, entry := range hosts {
		console.EmptyLine()
		console.Progress(i+1, len(hosts), entry.Name, entry.Host.Host)

		err := e.runOnHost(task, &entry.Host, globalVars, entry.Vars)
		results = append(results, Result{
			Host:    entry.Name,
			Success: err == nil,
			Error:   err,
		})

		if err != nil {
			console.Error("%s: %v", entry.Name, err)
		} else if !e.dryRun {
			console.Success("%s OK", entry.Name)
		}
	}

	return results
}

// runParallel 并发执行，按主机缓冲输出。
func (e *Executor) runParallel(task *config.Task, hosts []HostEntry, globalVars resolve.Vars) []Result {
	results := make([]Result, len(hosts))
	sem := make(chan struct{}, e.concurrency)
	var wg sync.WaitGroup

	for i, entry := range hosts {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, entry HostEntry) {
			defer wg.Done()
			defer func() { <-sem }()

			console.Progress(idx+1, len(hosts), entry.Name, entry.Host.Host)

			var buf bytes.Buffer
			console.SetOutput(&buf)

			defer func() {
				console.ResetOutput()

				if r := recover(); r != nil {
					results[idx] = Result{
						Host:    entry.Name,
						Success: false,
						Error:   fmt.Errorf("panic: %v", r),
					}
				}

				console.RawOutput(buf.String())
			}()

			console.Section(fmt.Sprintf("%s [%s]", entry.Name, entry.Host.Host))

			err := e.runOnHost(task, &entry.Host, globalVars, entry.Vars)

			results[idx] = Result{
				Host:    entry.Name,
				Success: err == nil,
				Error:   err,
			}

			if err != nil {
				console.Error("%s: %v", entry.Name, err)
			} else if !e.dryRun {
				console.Success("%s OK", entry.Name)
			}
		}(i, entry)
	}

	wg.Wait()
	return results
}

// runOnHost 对单台主机执行整条任务。
//
// 变量合并顺序：
//
//	global_vars -> 组 vars -> 主机 Extra -> CLI -D -> execId（运行时）
func (e *Executor) runOnHost(task *config.Task, host *config.Host,
	globalVars, hostVars resolve.Vars) error {

	vars := mergeVars(globalVars, hostVars)

	// CLI 变量优先级最高
	maps.Copy(vars, e.cliVars)

	// execId 由运行时注入，覆盖一切
	vars["execId"] = e.execID

	var client *ssh.Client

	if !e.dryRun {
		c, err := ssh.Connect(host, e.connectTimeout)
		if err != nil {
			return fmt.Errorf("connect %s: %w", host.Host, err)
		}
		client = c
		defer client.Close()
	}

	total := len(task.Steps)
	showStep := total > 1

	for i, step := range task.Steps {
		if i > 0 {
			console.EmptyLine()
		}
		if showStep {
			console.Step(i+1, total, step.Name)
		}

		// 条件判断
		if step.When != "" {
			ok, expanded, err := evalWhen(step.When, vars)
			if err != nil {
				return fmt.Errorf("step %s: eval when: %w", step.Name, err)
			}
			if !ok {
				console.Info("Skipped, condition not met: %s", expanded)
				continue
			}
		}

		act := e.actions.Get(step.Action)
		if act == nil {
			return fmt.Errorf("unknown action: %s", step.Action)
		}

		ctx := &action.Context{
			Host:   host,
			Vars:   vars,
			With:   step.With,
			Client: client,
			DryRun: e.dryRun,
		}

		if err := act.Execute(ctx); err != nil {
			if step.IgnoreErrors {
				console.Warn("step %s failed but ignored: %v", step.Name, err)
			} else {
				return fmt.Errorf("step %s: %w", step.Name, err)
			}
		}

		if step.Delay > 0 {
			console.Info("Waiting %ds...", step.Delay)
			if !e.dryRun {
				time.Sleep(time.Duration(step.Delay) * time.Second)
			}
		}
	}

	return nil
}

// runBatches 按批队列依次执行。
func (e *Executor) runBatches(task *config.Task, plan []Batch, globalVars resolve.Vars) []Result {
	var all []Result

	for i, batch := range plan {
		console.Section(fmt.Sprintf("Group %s, Batch %d/%d (%d hosts)",
			batch.Group, i+1, len(plan), len(batch.Hosts)))

		var results []Result
		if e.concurrency == 1 || e.dryRun {
			results = e.runSerial(task, batch.Hosts, globalVars)
		} else {
			results = e.runParallel(task, batch.Hosts, globalVars)
		}

		all = append(all, results...)

		if e.stopOnFailure {
			for _, r := range results {
				if !r.Success {
					console.Error("Batch failed, stopping remaining batches")
					return all
				}
			}
		}
	}

	return all
}

// mergeVars 合并变量池：global -> host，host 优先。
func mergeVars(globalVars, hostVars resolve.Vars) resolve.Vars {
	vars := make(resolve.Vars, len(globalVars)+len(hostVars))
	maps.Copy(vars, globalVars)
	maps.Copy(vars, hostVars)
	return vars
}

// printPlan 在 dry-run 时打印完整执行队列。
func printPlan(plan []Batch) {
	console.Section("Execution Plan")
	for i, batch := range plan {
		console.Info("Batch %d: group %s, %d hosts", i+1, batch.Group, len(batch.Hosts))
		for _, h := range batch.Hosts {
			console.Info("  - %s [%s]", h.Name, h.Host.Host)
		}
	}
}

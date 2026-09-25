package task

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"mestrap.com/taskssh/internal/action"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/log"
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
}

// NewExecutor 创建执行器。
//
// concurrency 为并发数，<=0 时按 1 处理。
// connectTimeout 为连接超时秒数，<=0 时使用默认值。
func NewExecutor(actions *action.Registry, concurrency int, connectTimeout int) *Executor {
	if concurrency <= 0 {
		concurrency = 1
	}
	if connectTimeout <= 0 {
		connectTimeout = 10
	}
	return &Executor{
		actions:        actions,
		concurrency:    concurrency,
		connectTimeout: time.Duration(connectTimeout) * time.Second,
	}
}

// Run 对一批主机执行任务。
func (e *Executor) Run(task *config.Task, hosts []HostEntry, globalVars resolve.Vars) []Result {
	log.Section("Start")

	if task.Description != "" {
		log.Info("Task: %s (%d steps)", task.Description, len(task.Steps))
	} else {
		log.Info("Steps: %d", len(task.Steps))
	}

	log.Info("Targets: %d", len(hosts))
	log.Info("Concurrency: %d", e.concurrency)

	if e.concurrency == 1 {
		return e.runSerial(task, hosts, globalVars)
	}
	return e.runParallel(task, hosts, globalVars)
}

// runSerial 串行执行，实时输出。
func (e *Executor) runSerial(task *config.Task, hosts []HostEntry, globalVars resolve.Vars) []Result {
	results := make([]Result, 0, len(hosts))

	for i, entry := range hosts {
		log.EmptyLine()
		log.Progress(i+1, len(hosts), entry.Name, entry.Host.Host)

		err := e.runOnHost(task, &entry.Host, globalVars, entry.Vars)
		results = append(results, Result{
			Host:    entry.Name,
			Success: err == nil,
			Error:   err,
		})

		if err != nil {
			log.Error("%s: %v", entry.Name, err)
		} else {
			log.Success("%s OK", entry.Name)
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

			log.Progress(idx+1, len(hosts), entry.Name, entry.Host.Host)

			var buf bytes.Buffer
			log.SetOutput(&buf)

			defer func() {
				log.ResetOutput()

				if r := recover(); r != nil {
					results[idx] = Result{
						Host:    entry.Name,
						Success: false,
						Error:   fmt.Errorf("panic: %v", r),
					}
				}

				log.RawOutput(buf.String())
			}()

			log.Section(fmt.Sprintf("%s [%s]", entry.Name, entry.Host.Host))

			err := e.runOnHost(task, &entry.Host, globalVars, entry.Vars)

			results[idx] = Result{
				Host:    entry.Name,
				Success: err == nil,
				Error:   err,
			}

			if err != nil {
				log.Error("%s: %v", entry.Name, err)
			} else {
				log.Success("%s OK", entry.Name)
			}
		}(i, entry)
	}

	wg.Wait()
	return results
}

// runOnHost 对单台主机执行整条任务。
//
// 连接一次，整条任务复用；任务结束后关闭。
func (e *Executor) runOnHost(task *config.Task, host *config.Host,
	globalVars, hostVars resolve.Vars) error {

	vars := mergeVars(globalVars, hostVars)

	client, err := ssh.Connect(host, e.connectTimeout)
	if err != nil {
		return fmt.Errorf("connect %s: %w", host.Host, err)
	}
	defer client.Close()

	total := len(task.Steps)
	showStep := total > 1

	for i, step := range task.Steps {
		if i > 0 {
			log.EmptyLine()
		}
		if showStep {
			log.Step(i+1, total, step.Name)
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
		}

		if err := act.Execute(ctx); err != nil {
			return fmt.Errorf("step %s: %w", step.Name, err)
		}

		if step.Delay > 0 {
			log.Info("Waiting %ds...", step.Delay)
			time.Sleep(time.Duration(step.Delay) * time.Second)
		}
	}

	return nil
}

// mergeVars 合并变量池：global -> host，host 优先。
func mergeVars(globalVars, hostVars resolve.Vars) resolve.Vars {
	vars := make(resolve.Vars, len(globalVars)+len(hostVars))
	for k, v := range globalVars {
		vars[k] = v
	}
	for k, v := range hostVars {
		vars[k] = v
	}
	return vars
}

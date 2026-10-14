package task

import (
	"context"
	"time"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
)

// Result 是单台主机的执行结果。
type Result struct {
	Host    string
	Success bool
	Error   error
}

// Executor 执行任务。
type Executor struct {
	concurrency    int
	connectTimeout time.Duration
	dryRun         bool
	cliVars        resolve.Vars
	execID         string
	serial         int
	minAvailable   int
	stopOnFailure  bool
}

// NewExecutor 创建执行器。
//
// cliVars 是 CLI -D 传入的变量，优先级最高，覆盖清单中的所有同名变量。
func NewExecutor(concurrency int, connectTimeout int,
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

// Run 对一批主机执行任务。
func (e *Executor) Run(ctx context.Context, task *config.Task, hosts []HostEntry, globalVars resolve.Vars) ([]Result, error) {

	e.execID = generateExecID()

	// 启用分批：serial > 0 或 minAvailable > 0
	if e.serial > 0 || e.minAvailable > 0 {
		plan, err := planBatches(hosts, e.serial, e.minAvailable)
		if err != nil {
			return nil, err
		}

		printPlan(plan)

		return e.runBatches(ctx, task, plan, globalVars), nil
	}

	if e.concurrency == 1 || e.dryRun {
		return e.runSerial(ctx, task, hosts, globalVars), nil
	}
	return e.runParallel(ctx, task, hosts, globalVars), nil
}

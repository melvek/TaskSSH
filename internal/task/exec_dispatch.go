package task

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/host"
	"mestrap.com/taskssh/internal/resolve"
)

// runSerial 串行执行，实时输出。
func (e *Executor) runSerial(ctx context.Context, task *config.Task, hosts []host.HostEntry, globalVars resolve.Vars) []Result {
	results := make([]Result, 0, len(hosts))

	for i, entry := range hosts {
		if ctx.Err() != nil {
			break
		}
		console.Progress(i+1, len(hosts), entry.Name, entry.Host.Host)

		err := e.runOnHost(ctx, task, entry, globalVars, console.DefaultLogger())
		results = append(results, Result{
			Host:    entry.Name,
			Success: err == nil,
			Error:   err,
		})

		// ctx 取消时停止后续主机
		if ctx.Err() != nil {
			break
		}
	}

	return results
}

// runParallel 并发执行，按主机缓冲输出。
func (e *Executor) runParallel(ctx context.Context, task *config.Task, hosts []host.HostEntry, globalVars resolve.Vars) []Result {
	results := make([]Result, len(hosts))
	sem := make(chan struct{}, e.concurrency)
	var wg sync.WaitGroup
	dispatched := 0

dispatch:
	for i, entry := range hosts {
		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			break dispatch
		}
		if ctx.Err() != nil {
			<-sem
			break
		}

		results[i].Host = entry.Name
		dispatched++
		wg.Add(1)

		go func(idx int, entry host.HostEntry) {
			defer wg.Done()
			defer func() { <-sem }()

			var buf bytes.Buffer
			logger := console.NewLogger(&buf)
			defer func() {
				if r := recover(); r != nil {
					results[idx] = Result{
						Host:    entry.Name,
						Success: false,
						Error:   fmt.Errorf("panic: %v", r),
					}
				}

				console.RawOutput(buf.String())
			}()

			logger.Progress(idx+1, len(hosts), entry.Name, entry.Host.Host)
			err := e.runOnHost(ctx, task, entry, globalVars, logger)

			results[idx] = Result{
				Host:    entry.Name,
				Success: err == nil,
				Error:   err,
			}
		}(i, entry)
	}

	wg.Wait()
	return results[:dispatched]
}

// runBatches 按批队列依次执行。
func (e *Executor) runBatches(ctx context.Context, task *config.Task, plan []Batch, globalVars resolve.Vars) []Result {
	var all []Result

	for i, batch := range plan {
		if ctx.Err() != nil {
			break
		}

		console.Output(console.LevelBrief, console.StyleNone, "")
		console.Section(fmt.Sprintf("Group %s, Batch %d/%d (%d hosts)",
			batch.Group, i+1, len(plan), len(batch.Hosts)))

		var results []Result
		if e.concurrency == 1 || e.dryRun {
			results = e.runSerial(ctx, task, batch.Hosts, globalVars)
		} else {
			results = e.runParallel(ctx, task, batch.Hosts, globalVars)
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

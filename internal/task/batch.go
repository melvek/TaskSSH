package task

import (
	"fmt"
	"strings"

	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/host"
)

// Batch 是一批待执行的主机，属于同一组
type Batch struct {
	Group string
	Hosts []host.HostEntry
}

// groupEntries 按 Name 的 "/" 前缀分组，返回组顺序和分组结果。
func groupEntries(entries []host.HostEntry) ([]string, map[string][]host.HostEntry) {
	order := []string{}
	groups := map[string][]host.HostEntry{}

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

// 计算执行分组
func planBatches(entries []host.HostEntry, serial, minAvailable int) ([]Batch, error) {
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
			end := min(start+batchSize, len(groupHosts))
			plan = append(plan, Batch{Group: gname, Hosts: groupHosts[start:end]})
		}
	}

	return plan, nil
}

// printPlan 打印完整执行队列
func printPlan(plan []Batch) {
	console.Default.Section("Execution Plan")
	for i, batch := range plan {
		console.Default.Output(console.LevelBrief, console.StyleNone,
			"Batch %d: group %s, %d hosts", i+1, batch.Group, len(batch.Hosts))
		for _, h := range batch.Hosts {
			console.Default.Output(console.LevelBrief, console.StyleNone,
				"  - %s [%s]", h.Name, h.Host.Host)
		}
	}
}

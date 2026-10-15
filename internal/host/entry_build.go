package host

import (
	"fmt"
	"maps"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
)

// buildEntry 从清单构建一个 HostEntry。
//
// 合并顺序（低 -> 高）：
//
//	global defaults -> 组 vars -> 主机字段
func buildEntry(groupName, hostName string, group *config.Group, inv *config.Inventory) HostEntry {
	host := group.Hosts[hostName]

	// 合并：组 vars 覆盖 defaults，主机再覆盖
	groupVars := group.Vars
	groupVars.Merge(&inv.Defaults)
	host.Merge(&groupVars)

	// 复制 Extra 到 Vars
	vars := make(resolve.Vars, len(host.Extra))
	maps.Copy(vars, host.Extra)

	return HostEntry{
		Name: fmt.Sprintf("%s/%s", groupName, hostName),
		Host: host,
		Vars: vars,
	}
}

// standaloneEntry 构建独立主机条目。
//
// 用于命令行直接给出的主机名或 IP，不在清单中。
func standaloneEntry(hostName string, inv *config.Inventory) HostEntry {
	host := config.Host{Host: hostName}
	host.Merge(&inv.Defaults)

	vars := make(resolve.Vars, len(host.Extra))
	maps.Copy(vars, host.Extra)

	return HostEntry{
		Name: hostName,
		Host: host,
		Vars: vars,
	}
}

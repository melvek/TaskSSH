package task

import (
	"fmt"
	"maps"
	"net"
	"sort"

	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
)

// resolveGroup 展开服务器组。
func resolveGroup(groupName string, group *config.Group, inv *config.Inventory) []HostEntry {
	var entries []HostEntry

	names := make([]string, 0, len(group.Hosts))
	for name := range group.Hosts {
		names = append(names, name)
	}
	sort.Strings(names)

	groupVars := group.Vars
	groupVars.Merge(&inv.Defaults)

	for _, name := range names {
		host := group.Hosts[name]
		host.Merge(&groupVars)

		vars := make(resolve.Vars, len(host.Extra))
		maps.Copy(vars, host.Extra)

		if ip := resolveHostIP(host.Host); ip != "" {
			host.Host = ip
		}

		entries = append(entries, HostEntry{
			Name: fmt.Sprintf("%s/%s", groupName, name),
			Host: host,
			Vars: vars,
		})
	}

	return entries
}

// resolveStandalone 解析独立主机。
func resolveStandalone(hostName string, inv *config.Inventory) HostEntry {
	host := config.Host{
		Host: hostName,
	}
	host.Merge(&inv.Defaults)

	vars := make(resolve.Vars, len(host.Extra))
	maps.Copy(vars, host.Extra)

	if ip := resolveHostIP(hostName); ip != "" {
		host.Host = ip
	}

	return HostEntry{
		Name: hostName,
		Host: host,
		Vars: vars,
	}
}

// resolveHostIP 把主机名解析为 IP。
func resolveHostIP(host string) string {
	if host == "" {
		return ""
	}

	if net.ParseIP(host) != nil {
		return host
	}

	ips, err := net.LookupHost(host)
	if err != nil || len(ips) == 0 {
		return ""
	}

	return ips[0]
}

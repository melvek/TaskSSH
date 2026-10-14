package config

import "fmt"

// reservedVars 是 TaskSSH 运行时注入的变量名。
//
// 清单中不得定义这些名字，CLI -D 也不得覆盖。
var reservedVars = map[string]bool{
	"date":       true,
	"exec_id":    true,
	"host_name":  true,
	"group_name": true,
	"host_addr":  true,
	"host_port":  true,
	"host_user":  true,
}

// IsReservedVar 判断变量名是否为保留变量。
func IsReservedVar(name string) bool {
	return reservedVars[name]
}

// validateReservedVars 校验清单中是否定义了保留变量。
func validateReservedVars(inv *Inventory) error {
	if err := checkExtra("defaults", inv.Defaults.Extra); err != nil {
		return err
	}
	for groupName, group := range inv.Servers {
		scope := fmt.Sprintf("servers.%s.vars", groupName)
		if err := checkExtra(scope, group.Vars.Extra); err != nil {
			return err
		}
		for hostName, host := range group.Hosts {
			scope := fmt.Sprintf("servers.%s.hosts.%s", groupName, hostName)
			if err := checkExtra(scope, host.Extra); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkExtra(scope string, extra map[string]any) error {
	for key := range extra {
		if reservedVars[key] {
			return fmt.Errorf("%s: variable name %q is reserved by TaskSSH，please use a different name", scope, key)
		}
	}
	return nil
}

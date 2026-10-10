package config

import "fmt"

// includedFile 是被 include 的文件及其来源路径。
type includedFile struct {
	path string
	inv  *Inventory
}

// mergeInventories 合并主文件与 include 的子文件。
//
// 合并规则：
//   - global_vars：主文件优先，其次后 include，最后先 include
//   - servers：按组名合并，同名报错
//   - tasks：按任务名合并，同名报错
func mergeInventories(
	mainPath string,
	mainInv *Inventory,
	subs []includedFile,
) (*Inventory, error) {
	result := &Inventory{
		GlobalVars: mainInv.GlobalVars,
		Servers:    make(map[string]Group),
		Tasks:      make(map[string]Task),
	}

	// 记录每个组/任务的来源文件，用于冲突报错
	groupSource := make(map[string]string)
	taskSource := make(map[string]string)

	// 先合并所有 include 文件
	for _, sub := range subs {
		for name, g := range sub.inv.Servers {
			if src, ok := groupSource[name]; ok {
				return nil, fmt.Errorf(
					"duplicate group %q defined in:\n  - %s\n  - %s",
					name, src, sub.path)
			}
			result.Servers[name] = g
			groupSource[name] = sub.path
		}
		for name, t := range sub.inv.Tasks {
			if src, ok := taskSource[name]; ok {
				return nil, fmt.Errorf(
					"duplicate task %q defined in:\n  - %s\n  - %s",
					name, src, sub.path)
			}
			result.Tasks[name] = t
			taskSource[name] = sub.path
		}
	}

	// 再合并主文件的 servers / tasks
	for name, g := range mainInv.Servers {
		if src, ok := groupSource[name]; ok {
			return nil, fmt.Errorf(
				"duplicate group %q defined in:\n  - %s\n  - %s",
				name, src, mainPath)
		}
		result.Servers[name] = g
	}
	for name, t := range mainInv.Tasks {
		if src, ok := taskSource[name]; ok {
			return nil, fmt.Errorf(
				"duplicate task %q defined in:\n  - %s\n  - %s",
				name, src, mainPath)
		}
		result.Tasks[name] = t
	}

	// global_vars：从后往前 Merge
	//
	// Host.Merge 的语义是「source 补充 target 的空白」，
	// 所以从后往前 Merge 得到优先级：
	//   主文件 > subs[last] > ... > subs[0]
	for i := len(subs) - 1; i >= 0; i-- {
		result.GlobalVars.Merge(&subs[i].inv.GlobalVars)
	}

	return result, nil
}
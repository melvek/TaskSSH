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
//   - servers：按组名合并，同名报错
//   - tasks：按任务名合并，同名报错
//   - defaults：只保留主文件的，子文件的 defaults 已在加载时下沉
func mergeInventories(mainPath string, mainInv *Inventory, subs []includedFile) (*Inventory, error) {
	result := &Inventory{
		Defaults: mainInv.Defaults,
		Servers:  make(map[string]Group),
		Tasks:    make(map[string]Task),
	}

	groupSource := make(map[string]string)
	taskSource := make(map[string]string)

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

	return result, nil
}

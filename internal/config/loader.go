package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/fileutil"
)

const DefaultInventory = "inventory.yaml"

// 加载清单文件
func Load(path string) (*Inventory, error) {
	// 如果清单未指定清单文件，则使用默认的路径
	if path == "" {
		path = DefaultInventory
	}

	path = fileutil.ExpandHome(path)

	// 指定的清单文件不存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// 清单文件并不是必需的
		return finalizeInventory(&Inventory{})
	}

	console.Default.Debug("loading invertory file: %s ...", path)
	inv, err := readInventoryFile(path)
	if err != nil {
		return nil, err
	}

	// 没有 include 文件时，处理参数合并
	if len(inv.Include) == 0 {
		return finalizeInventory(inv)
	}

	// 展开 include 为文件列表
	baseDir := filepath.Dir(path)
	files, err := expandIncludes(inv.Include, baseDir)
	if err != nil {
		return nil, err
	}

	// 依次加载子文件
	subs := make([]includedFile, 0, len(files))
	for _, f := range files {
		console.Default.Debug("loading include invertory file: %s ...", f)
		sub, err := readInventoryFile(f)
		if err != nil {
			return nil, err
		}
		if len(sub.Include) > 0 {
			return nil, fmt.Errorf(
				"include is not allowed in included file: %s\n"+
					"       include directive is only supported in the main inventory file.\n"+
					"       To organize nested content, list files explicitly in the main inventory:\n"+
					"         include:\n"+
					"           - servers/prod/web.yaml\n"+
					"           - servers/prod/db.yaml",
				f)
		}

		applyDefaultsToGroups(sub)

		subs = append(subs, includedFile{path: f, inv: sub})
	}

	// 合并
	result, err := mergeInventories(path, inv, subs)
	if err != nil {
		return nil, err
	}

	return finalizeInventory(result)
}

// 读取单个清单文件，不含 include 处理
func readInventoryFile(path string) (*Inventory, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("inventory file not found: %s\n"+
				"       create one or specify with -i / --inventory", path)
		}
		return nil, fmt.Errorf("read inventory %s: %w", path, err)
	}

	var inv Inventory
	if err := yaml.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("parse inventory %s: %w", path, err)
	}

	if inv.Defaults.Extra == nil {
		inv.Defaults.Extra = make(map[string]any)
	}

	return &inv, nil
}

func finalizeInventory(inv *Inventory) (*Inventory, error) {
	if inv.Servers == nil {
		inv.Servers = make(map[string]Group)
	}
	if inv.Tasks == nil {
		inv.Tasks = make(map[string]Task)
	}
	if inv.Defaults.Extra == nil {
		inv.Defaults.Extra = make(map[string]any)
	}

	if err := validateHostSelectorNames(inv); err != nil {
		return nil, err
	}
	if err := validateReservedVars(inv); err != nil {
		return nil, err
	}

	inv.Defaults.Extra["date"] = time.Now().Format("20060102")
	inv.Include = nil

	return inv, nil
}

func validateHostSelectorNames(inv *Inventory) error {
	for groupName, group := range inv.Servers {
		if strings.ContainsRune(groupName, '/') {
			return fmt.Errorf("invalid group name %q: '/' is reserved for host selectors", groupName)
		}
		for hostName := range group.Hosts {
			if strings.ContainsRune(hostName, '/') {
				return fmt.Errorf("invalid host name %q in group %q: '/' is reserved for host selectors", hostName, groupName)
			}
		}
	}
	return nil
}

// applyDefaultsToGroups 把 inv.Defaults 下沉到 inv.Servers 里每个 group 的 Vars，
// 然后清空 inv.Defaults。
//
// 用于 include 的子文件：子文件的 defaults 只作用于该文件的 servers，
// 不会污染其他文件。
//
// 合并方向：group.Vars 优先，inv.Defaults 补充空白。
func applyDefaultsToGroups(inv *Inventory) {
	for name, group := range inv.Servers {
		groupVars := group.Vars
		groupVars.Merge(&inv.Defaults)
		group.Vars = groupVars
		inv.Servers[name] = group
	}
	inv.Defaults = Host{Extra: map[string]any{}}
}

// expandIncludes 展开 include 模式列表
func expandIncludes(patterns []string, baseDir string) ([]string, error) {
	var files []string

	for _, pattern := range patterns {
		expanded, err := expandIncludePattern(pattern, baseDir)
		if err != nil {
			return nil, err
		}
		files = append(files, expanded...)
	}

	return files, nil
}

// expandIncludePattern 展开单个 include 模式。
//
// 支持三种形式：
//   - 精确路径：servers/prod.yaml
//   - 目录：servers/  → servers/*.yaml
//   - glob：servers/*.yaml
//
// 相对路径基于 baseDir。glob 结果按文件名排序。
func expandIncludePattern(pattern, baseDir string) ([]string, error) {
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(baseDir, pattern)
	}

	// 目录 → 追加 /*.yaml
	if info, err := os.Stat(pattern); err == nil && info.IsDir() {
		pattern = filepath.Join(pattern, "*.yaml")
	}

	// 无 glob 元字符：精确路径
	if !strings.ContainsAny(pattern, "*?[") {
		if _, err := os.Stat(pattern); err != nil {
			return nil, fmt.Errorf("include file not found: %s", pattern)
		}
		return []string{pattern}, nil
	}

	// glob 展开
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid include pattern %s: %w", pattern, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("include pattern matched no files: %s", pattern)
	}

	sort.Strings(matches)
	return matches, nil
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Load 从文件加载清单。
//
// 如果主文件含 include 段，则加载被 include 的文件并合并。
// 子文件不允许再有 include 段。
func Load(path string) (*Inventory, error) {
	inv, err := readInventoryFile(path)
	if err != nil {
		return nil, err
	}

	// 无 include：直接收尾
	if len(inv.Include) == 0 {
		return finalizeInventory(inv), nil
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
		subs = append(subs, includedFile{path: f, inv: sub})
	}

	// 合并
	result, err := mergeInventories(path, inv, subs)
	if err != nil {
		return nil, err
	}

	return finalizeInventory(result), nil
}

// readInventoryFile 读取单个清单文件，不含 include 处理。
func readInventoryFile(path string) (*Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read inventory %s: %w", path, err)
	}

	var inv Inventory
	if err := yaml.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("parse inventory %s: %w", path, err)
	}

	if inv.GlobalVars.Extra == nil {
		inv.GlobalVars.Extra = make(map[string]interface{})
	}

	return &inv, nil
}

// finalizeInventory 收尾：初始化 map、注入内置变量、清理 Include 字段。
func finalizeInventory(inv *Inventory) *Inventory {
	if inv.Servers == nil {
		inv.Servers = make(map[string]Group)
	}
	if inv.Tasks == nil {
		inv.Tasks = make(map[string]Task)
	}
	if inv.GlobalVars.Extra == nil {
		inv.GlobalVars.Extra = make(map[string]interface{})
	}

	inv.GlobalVars.Extra["date"] = time.Now().Format("20060102")
	inv.Include = nil

	return inv
}

// expandIncludes 展开 include 模式列表。
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

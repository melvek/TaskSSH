package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Load 从文件加载清单。
func Load(path string) (*Inventory, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read inventory %s: %w", path, err)
	}

	var inv Inventory
	if err := yaml.Unmarshal(data, &inv); err != nil {
		return nil, fmt.Errorf("parse inventory %s: %w", path, err)
	}

	// 初始化 Extra
	if inv.GlobalVars.Extra == nil {
		inv.GlobalVars.Extra = make(map[string]interface{})
	}

	// 注入 date 变量
	inv.GlobalVars.Extra["date"] = time.Now().Format("20060102")

	return &inv, nil
}

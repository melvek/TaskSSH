package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Host 描述主机配置。
// 同时用于 global_vars、组的 vars、以及具体主机。
type Host struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	IdentityFile string `yaml:"identity_file,omitempty"`
	Passphrase   string `yaml:"passphrase,omitempty"`

	// Extra 存放未定义的字段，用于变量替换
	Extra map[string]interface{} `yaml:",inline"`
}

// UnmarshalYAML 支持主机的两种写法：
//   - 简写：web_1: 1.2.3.4
//   - 完整：web_2: { host: 1.2.3.5, port: 2222 }
func (h *Host) UnmarshalYAML(value *yaml.Node) error {
	// 字符串形式：直接当 host
	if value.Kind == yaml.ScalarNode {
		h.Host = value.Value
		return nil
	}

	// 对象形式：正常解析
	type rawHost Host
	var raw rawHost
	if err := value.Decode(&raw); err != nil {
		return fmt.Errorf("parse host: %w", err)
	}
	*h = Host(raw)

	// 初始化 Extra
	if h.Extra == nil {
		h.Extra = make(map[string]interface{})
	}
	return nil
}

// Merge 将另一个 Host 的值合并进来。
// 规则：目标已有值则跳过，不覆盖。
func (h *Host) Merge(source *Host) {
	if source == nil {
		return
	}
	if source.Port != 0 && h.Port == 0 {
		h.Port = source.Port
	}
	if source.Username != "" && h.Username == "" {
		h.Username = source.Username
	}
	if source.Password != "" && h.Password == "" {
		h.Password = source.Password
	}
	if source.IdentityFile != "" && h.IdentityFile == "" {
		h.IdentityFile = source.IdentityFile
	}
	if source.Passphrase != "" && h.Passphrase == "" {
		h.Passphrase = source.Passphrase
	}
	if h.Extra == nil {
		h.Extra = make(map[string]interface{})
	}
	for k, v := range source.Extra {
		if _, ok := h.Extra[k]; !ok {
			h.Extra[k] = v
		}
	}
}

// Package host 提供主机解析：从 Inventory 解析出目标主机列表。
//
// 输入支持 pattern 语法（g: / h: / ! 前缀），输出 []HostEntry。
// 每个 HostEntry 包含主机的连接信息和变量池。
package host

import (
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/resolve"
)

// Overrides 是 CLI 覆盖参数。
type Overrides struct {
	Port         int
	Username     string
	Password     string
	IdentityFile string
}

// HostEntry 是目标主机的一项。
type HostEntry struct {
	Name string
	Host config.Host
	Vars resolve.Vars
}

// ResolveHosts 解析目标主机 pattern，返回匹配的主机列表。
//
// pattern 支持空格或逗号分隔的多个子 pattern。
//
// 每个子 pattern：
//
//	prod                    精确匹配组名
//	prod/web1               精确匹配 "组名/主机名"
//	web1                    精确匹配组内主机名（多组同名时报错）
//	192.168.1.10            独立主机或 IP
//	g:web*                  glob 匹配组名
//	h:web*                  glob 匹配组内主机名
//	web*                    默认：同时匹配组名和主机名，去重合并
//	!g:web*                 排除：组名匹配 web* 的组
//	!h:web1                 排除：主机名匹配 web1 的主机
//
// 无 include 时默认 g:*（所有组）。exclude 匹配错误不中断。
func ResolveHosts(pattern string, inv *config.Inventory, ov Overrides) ([]HostEntry, error) {
	r := newResolver(inv)
	return r.resolve(pattern, ov)
}

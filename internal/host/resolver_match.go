package host

import (
	"fmt"
	"net"
	"path"
	"strings"
)

// byGroup 匹配组名。
//
// 空 pattern（"g:" 后无内容）视为语法错误；
// path.Match 出错视为语法错误；
// strict=true 时匹配不到返回 error，strict=false 时返回 nil, nil。
func (r *resolver) byGroup(pattern string, strict bool) ([]HostEntry, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty group pattern after 'g:'")
	}

	// 精确匹配
	if idxs, ok := r.groups[pattern]; ok {
		return r.entriesAt(idxs), nil
	}

	// glob 匹配
	if isGlobPattern(pattern) {
		entries, err := r.matchGroups(pattern)
		if err != nil {
			return nil, err
		}
		if len(entries) > 0 {
			return entries, nil
		}
		if strict {
			return nil, fmt.Errorf("no group matched: %q", pattern)
		}
		return nil, nil
	}

	if strict {
		return nil, fmt.Errorf("no group matched: %q", pattern)
	}
	return nil, nil
}

// byHost 匹配所有组内主机名。
//
// 空 pattern（"h:" 后无内容）视为语法错误；
// path.Match 出错视为语法错误；
// strict=true 时匹配不到返回 error，strict=false 时返回 nil, nil。
func (r *resolver) byHost(pattern string, strict bool) ([]HostEntry, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty host pattern after 'h:'")
	}

	// 精确匹配（走 O(1) 索引）
	if !isGlobPattern(pattern) {
		entries := r.findAllHostsByName(pattern)
		if len(entries) > 0 {
			return entries, nil
		}
		if strict {
			return nil, fmt.Errorf("no host matched: %q", pattern)
		}
		return nil, nil
	}

	// glob 匹配
	entries, err := r.matchHosts(pattern)
	if err != nil {
		return nil, err
	}
	if len(entries) > 0 {
		return entries, nil
	}
	if strict {
		return nil, fmt.Errorf("no host matched: %q", pattern)
	}
	return nil, nil
}

// byDefault 无前缀的默认匹配。
//
// strict 控制匹配不到时的行为：
//   - strict=true（& / ! 操作数）：匹配不到直接报错，绝不回退到独立主机，
//     避免用户打错字时减法/交集静默失效；
//   - strict=false（普通 token）：匹配不到且名字是合法主机/IP/域名时，
//     视为独立主机（ad-hoc 主机）；名字含非法字符或 glob 元字符时报错。
//
// 匹配顺序：
//  1. 精确匹配组名
//  2. 含 "/"：组/主机组合匹配（两段各自支持精确与 glob）
//  3. 精确匹配组内主机名（跨组返回所有同名主机）
//  4. glob 匹配组名 + 主机名，去重合并
//  5. 以上都不匹配 → strict=false 且名字合法 → 独立主机；否则报错
func (r *resolver) byDefault(name string, strict bool) ([]HostEntry, error) {
	// 1. 精确匹配组名
	if idxs, ok := r.groups[name]; ok {
		return r.entriesAt(idxs), nil
	}

	// 2. 含 "/"：组/主机组合匹配
	if groupPat, hostPat, found := strings.Cut(name, "/"); found {
		// groupPat/hostPat 任一侧为空 → 畸形
		if groupPat == "" || hostPat == "" {
			return nil, fmt.Errorf("malformed host pattern %q: empty group or host name", name)
		}
		// 多段斜杠（主机名中含 "/"）→ 畸形
		if strings.ContainsRune(hostPat, '/') {
			return nil, fmt.Errorf("malformed host pattern %q: too many slashes", name)
		}

		groupIsGlob := isGlobPattern(groupPat)
		hostIsGlob := isGlobPattern(hostPat)

		// 先校验 glob 语法（避免进入循环后才发现）
		if groupIsGlob {
			if _, err := path.Match(groupPat, ""); err != nil {
				return nil, fmt.Errorf("invalid group pattern %q: %w", groupPat, err)
			}
		}
		if hostIsGlob {
			if _, err := path.Match(hostPat, ""); err != nil {
				return nil, fmt.Errorf("invalid host pattern %q: %w", hostPat, err)
			}
		}

		// 精确组名 + 精确主机名：组存在但主机不存在 → 明确报错
		if !groupIsGlob && !hostIsGlob {
			group, groupExists := r.inv.Servers[groupPat]
			if !groupExists {
				// 这个名字同时是"组/主机"格式，但任一侧都不是 glob 且组不存在：
				// 不能当作独立主机（名字里带 "/"），也不能报"组不存在"（因为
				// 用户可能想表达的是别的）——直接报整体无匹配。
				return nil, fmt.Errorf("no hosts matched: %q", name)
			}
			if _, hostExists := group.Hosts[hostPat]; hostExists {
				// 组与主机都存在却没命中索引 → 内部一致性错误
				if idx, ok := r.byName[name]; ok {
					return []HostEntry{r.entries[idx]}, nil
				}
				return nil, fmt.Errorf("internal error: host %q indexed but not found in entries", name)
			}
			return nil, fmt.Errorf("host %q not found in group %q", hostPat, groupPat)
		}

		// glob 匹配
		entries, err := r.matchGroupHost(groupPat, hostPat)
		if err != nil {
			return nil, err
		}
		if len(entries) > 0 {
			return entries, nil
		}
		// glob 模式匹配不到一律报错（含斜杠不能当独立主机）
		return nil, fmt.Errorf("no hosts matched: %q", name)
	}

	// 3. 精确匹配组内主机名（跨组返回所有）
	if all := r.findAllHostsByName(name); len(all) > 0 {
		return all, nil
	}

	// 3.5 IP 字面量（含 [::1]、[fe80::1%eth0] 这类带方括号的写法）优先按
	// 独立主机处理。它们的 "[" 会被 isGlobPattern 误判为 glob 元字符，
	// 若不在此提前拦截，[::1] 会被当成非法的 glob 模式而遭拒。
	if parseStandaloneIP(name) != nil {
		if strict {
			return nil, fmt.Errorf("no hosts matched: %q", name)
		}
		return []HostEntry{standaloneEntry(name, r.inv)}, nil
	}

	// 4. glob 匹配组名 + 主机名，去重合并
	if isGlobPattern(name) {
		// 先校验语法
		if _, err := path.Match(name, ""); err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", name, err)
		}
		groupHits, err := r.matchGroups(name)
		if err != nil {
			return nil, err
		}
		hostHits, err := r.matchHosts(name)
		if err != nil {
			return nil, err
		}
		all := append(groupHits, hostHits...)
		if len(all) > 0 {
			return dedupeEntries(all), nil
		}
		// glob 匹配不到 → 报错（不能当作名字含 * 的独立主机）
		return nil, fmt.Errorf("no hosts matched: %q", name)
	}

	// 5. 独立主机
	if strict {
		return nil, fmt.Errorf("no hosts matched: %q", name)
	}
	if err := validateStandaloneName(name); err != nil {
		return nil, fmt.Errorf("invalid host name %q: %w", name, err)
	}
	return []HostEntry{standaloneEntry(name, r.inv)}, nil
}

// matchGroups 返回所有组名匹配 pattern 的条目。
//
// 调用方负责先校验 pattern 语法；path.Match 出错仅在 pattern 非法时触发，
// 此处仍返回错误以防意外。
func (r *resolver) matchGroups(pattern string) ([]HostEntry, error) {
	all := make([]HostEntry, 0, len(r.entries))
	for _, groupName := range r.groupOrder {
		ok, err := path.Match(pattern, groupName)
		if err != nil {
			return nil, fmt.Errorf("invalid group pattern %q: %w", pattern, err)
		}
		if ok {
			all = append(all, r.entriesAt(r.groups[groupName])...)
		}
	}
	return all, nil
}

// matchHosts 返回所有主机名匹配 pattern 的条目。
//
// 平铺遍历 entries（其顺序即 groupOrder x hostOrder），借 hostOf 直接取主机名，
// 全程无字符串拼接、无哈希查表。
func (r *resolver) matchHosts(pattern string) ([]HostEntry, error) {
	all := make([]HostEntry, 0, len(r.entries))
	for i := range r.entries {
		ok, err := path.Match(pattern, r.hostOf[i])
		if err != nil {
			return nil, fmt.Errorf("invalid host pattern %q: %w", pattern, err)
		}
		if ok {
			all = append(all, r.entries[i])
		}
	}
	return all, nil
}

// matchGroupHost 匹配 "组 pattern/主机 pattern"。
//
// 两部分各自支持精确和 glob：
//
//	prod/web1       精确组名 + 精确主机名
//	prod/web*       精确组名 + 主机 glob
//	prod*/web1      组 glob + 精确主机名
//	prod*/web*      组 glob + 主机 glob
//
// 匹配不到返回 nil, nil；pattern 语法错返回 error。
func (r *resolver) matchGroupHost(groupPat, hostPat string) ([]HostEntry, error) {
	all := make([]HostEntry, 0, len(r.entries))
	for _, groupName := range r.groupOrder {
		ok, err := path.Match(groupPat, groupName)
		if err != nil {
			return nil, fmt.Errorf("invalid group pattern %q: %w", groupPat, err)
		}
		if !ok {
			continue
		}
		// 直接用该组的 entries 下标，主机名从 hostOf 取，无拼接无查表
		for _, idx := range r.groups[groupName] {
			ok, err := path.Match(hostPat, r.hostOf[idx])
			if err != nil {
				return nil, fmt.Errorf("invalid host pattern %q: %w", hostPat, err)
			}
			if ok {
				all = append(all, r.entries[idx])
			}
		}
	}
	return all, nil
}

// isGlobPattern 判断字符串是否含 glob 元字符。
func isGlobPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// ---------------------------------------------------------------------------
// 独立主机名校验
// ---------------------------------------------------------------------------

// standaloneNameMaxLen 独立主机名最大长度。
// DNS 全域名限制 253 字符；IPv6 地址最长 39 字符（不含 zone ID）。
// 取 253 与 DNS 规范对齐。
const standaloneNameMaxLen = 253

// validateStandaloneName 校验独立主机名/域名/IP 字面量是否合法。
//
// 允许以下三类：
//  1. IPv4 点分十进制（如 192.168.1.1，允许前导零，按 inet_aton 语义）
//  2. IPv6 字面量（含 zone ID 如 fe80::1%eth0；自动剥离 [] 包裹）
//  3. 主机名 / 域名（RFC 952/1123 基础上放宽下划线，兼容 SRV/k8s 命名；
//     每段 label 1-63 字符，字母数字开头结尾，中间允许 - _）
//
// 明确拒绝：
//   - 空串
//   - 超过 253 字符
//   - glob 元字符 * ? [
//   - shell 元字符 <space> ' " ` $ & | ; < > ( ) { } ! \n \r \t
//   - # @ : / \ = + 等非主机名字符（冒号会与 host:port 语义冲突；@ 属于 user@host）
//   - 前导/尾随连字符或下划线、连续点、连续下划线/连字符
//   - 全数字+点、含冒号但非合法 IP 的"像 IP 却不是"的字符串
func validateStandaloneName(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	if len(name) > standaloneNameMaxLen {
		return fmt.Errorf("name too long (%d > %d)", len(name), standaloneNameMaxLen)
	}

	// 1) 尝试 IP 解析（含 IPv6 zone ID / [] 包裹）
	if ip := parseStandaloneIP(name); ip != nil {
		return nil
	}

	// 1b) 长得像 IPv4（含点且只含数字与点）但 net.ParseIP 失败 → octet 非法
	if looksLikeIPv4(name) {
		return fmt.Errorf("invalid IPv4 address %q", name)
	}
	// 1c) 含冒号（IPv6 特征）但解析失败 → 非法 IPv6
	if strings.ContainsRune(name, ':') {
		return fmt.Errorf("invalid IPv6 address %q", name)
	}

	// 2) 主机名/域名校验
	return validateHostname(name)
}

// looksLikeIPv4 判断 name 是否是纯数字+点的形如 IPv4 的字符串。
//
// 要求至少 3 个点（4 段 octet），且所有非点字符都是数字；
// 单点/双点的数字串（如 "1234"、"1.2"）仍当作主机名/数字 label 放行。
func looksLikeIPv4(name string) bool {
	dots := 0
	for _, c := range name {
		if c == '.' {
			dots++
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return dots >= 3
}

// parseStandaloneIP 尝试把 name 解析为 IP 地址。
//
// 支持：
//   - 纯 IPv4 / IPv6 字面量
//   - 带 zone ID 的 link-local IPv6（fe80::1%eth0）
//   - [] 包裹的 IPv6（[::1]、[fe80::1%eth0]）
//
// zone ID 仅允许字母、数字、下划线、点、连字符（常见网卡名形式）。
// 返回 nil 表示不是合法 IP 字面量（可能是主机名）。
func parseStandaloneIP(name string) net.IP {
	s := name

	// 剥离 [] 包裹
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") && len(s) >= 3 {
		s = s[1 : len(s)-1]
	}

	// IPv6 zone ID：% 后的部分不影响 IP 合法性，先校验 zone 再剥离
	if i := strings.IndexByte(s, '%'); i >= 0 {
		zone := s[i+1:]
		if zone == "" || !isValidZoneID(zone) {
			return nil
		}
		s = s[:i]
	}
	if s == "" {
		return nil
	}

	ip := net.ParseIP(s)
	return ip
}

// isValidZoneID zone ID 仅允许 [A-Za-z0-9_.-]（Linux 网卡名 / Windows ifindex 均满足）。
func isValidZoneID(zone string) bool {
	for _, c := range zone {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case c == '_' || c == '.' || c == '-':
		default:
			return false
		}
	}
	return true
}

// validateHostname 校验主机名/域名格式。
//
// 规则汇总：
//   - 总长度 ≤ 253 字符（已在外层检查）
//   - 支持零或多个点分隔的 label（允许单段短名、允许多段 FQDN）
//   - label 长度 1-63
//   - label 首尾必须为字母或数字
//   - label 中间允许字母、数字、连字符、下划线
//   - 禁止连续点（空 label）
//   - 禁止前导点、尾随点
func validateHostname(name string) error {
	if name == "" {
		return fmt.Errorf("empty name")
	}
	// 先扫一遍非法字符
	for _, c := range name {
		if !isHostnameChar(c) {
			if c == '*' || c == '?' || c == '[' {
				return fmt.Errorf("glob characters are not allowed in host names (found %q)", c)
			}
			return fmt.Errorf("invalid character %q in host name", c)
		}
	}

	if name[0] == '.' || name[len(name)-1] == '.' {
		return fmt.Errorf("host name cannot start or end with '.'")
	}
	if strings.Contains(name, "..") {
		return fmt.Errorf("consecutive dots in host name")
	}

	labels := strings.Split(name, ".")
	for _, label := range labels {
		if err := validateHostLabel(label); err != nil {
			return err
		}
	}
	return nil
}

// hostLabelMaxLen DNS label 最大长度。
const hostLabelMaxLen = 63

// validateHostLabel 校验单个 hostname label。
func validateHostLabel(label string) error {
	if label == "" {
		return fmt.Errorf("empty label in host name")
	}
	if len(label) > hostLabelMaxLen {
		return fmt.Errorf("host label %q too long (%d > %d)", label, len(label), hostLabelMaxLen)
	}
	first := label[0]
	last := label[len(label)-1]
	if !isAlnum(rune(first)) {
		return fmt.Errorf("host label %q must start with a letter or digit", label)
	}
	if !isAlnum(rune(last)) {
		return fmt.Errorf("host label %q must end with a letter or digit", label)
	}
	// 首尾已校验是 alnum，中间允许 alnum / - / _
	// 注意：单字符 label（如 "a"）时区间为空，必须用 i < len(label)-1 循环，
	// 不能写 label[1:len(label)-1]（len==1 时是 [1:0]，会 panic）。
	for i := 1; i < len(label)-1; i++ {
		c := rune(label[i])
		if !isHostnameChar(c) {
			return fmt.Errorf("invalid character %q in host label %q", c, label)
		}
	}
	return nil
}

// isHostnameChar 主机名字符集：字母、数字、连字符、下划线、点。
//
// 点用于分隔 label，不在单个 label 校验内调用。
func isHostnameChar(c rune) bool {
	switch {
	case c >= 'a' && c <= 'z':
	case c >= 'A' && c <= 'Z':
	case c >= '0' && c <= '9':
	case c == '-' || c == '_' || c == '.':
	default:
		return false
	}
	return true
}

// isAlnum 判断 c 是否为字母或数字。
func isAlnum(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

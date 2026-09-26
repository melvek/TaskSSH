package task

import (
	"fmt"
	"strconv"
	"strings"
)

// ExpandHostPattern 展开主机模式。
//
// 支持：
//
//	web[1-3]              → web1 web2 web3
//	web[01-03]            → web01 web02 web03
//	web[a,b,c]            → weba webb webc
//	web[1-3,5,7-9]        → web1..web3 web5 web7..web9
//	web[a-c]              → weba webb webc
//	web[1-2].dc[x,y].com  → web1.dcx.com web1.dcy.com web2.dcx.com web2.dcy.com
//
// 不包含 [] 的模式原样返回单个元素。
func ExpandHostPattern(pattern string) ([]string, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return nil, nil
	}

	// 没有 []，直接返回
	if !strings.Contains(pattern, "[") {
		return []string{pattern}, nil
	}

	// 找到第一对 []
	start := strings.Index(pattern, "[")
	end := strings.Index(pattern[start:], "]")
	if end < 0 {
		return nil, fmt.Errorf("unclosed bracket in pattern: %s", pattern)
	}
	end += start

	prefix := pattern[:start]
	body := pattern[start+1 : end]
	suffix := pattern[end+1:]

	// 展开 body
	items, err := expandBracketBody(body)
	if err != nil {
		return nil, fmt.Errorf("pattern %q: %w", pattern, err)
	}

	// 递归展开 suffix
	var result []string
	for _, item := range items {
		combined := prefix + item + suffix
		sub, err := ExpandHostPattern(combined)
		if err != nil {
			return nil, err
		}
		result = append(result, sub...)
	}

	return result, nil
}

// expandBracketBody 展开 [] 内的内容。
//
// 支持逗号分隔的多个项，每项可以是：
//   - 单个值：a
//   - 范围：1-10
//   - 带前导零的范围：01-03
//   - 字符范围：a-c
func expandBracketBody(body string) ([]string, error) {
	var result []string

	for _, part := range strings.Split(body, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 范围：1-10
		if strings.Contains(part, "-") {
			items, err := expandRange(part)
			if err != nil {
				return nil, err
			}
			result = append(result, items...)
			continue
		}

		// 单值
		result = append(result, part)
	}

	return result, nil
}

// expandRange 展开范围表达式，如 1-10、01-03、a-c。
func expandRange(expr string) ([]string, error) {
	parts := strings.SplitN(expr, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range: %q", expr)
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	// 数字范围
	startNum, errS := strconv.Atoi(startStr)
	endNum, errE := strconv.Atoi(endStr)

	if errS == nil && errE == nil {
		return expandNumericRange(startStr, endStr, startNum, endNum)
	}

	// 字符范围（单字符）
	if len(startStr) == 1 && len(endStr) == 1 {
		s := startStr[0]
		e := endStr[0]
		if s > e {
			return nil, fmt.Errorf("invalid range: %q (start > end)", expr)
		}
		var result []string
		for c := s; c <= e; c++ {
			result = append(result, string(c))
		}
		return result, nil
	}

	return nil, fmt.Errorf("invalid range: %q", expr)
}

// expandNumericRange 展开数字范围。
//
// 保留前导零的位数：
//
//	1-10  → 1, 2, ..., 10
//	01-10 → 01, 02, ..., 10
func expandNumericRange(startStr, endStr string, start, end int) ([]string, error) {
	if start > end {
		return nil, fmt.Errorf("invalid range: %d-%d (start > end)", start, end)
	}

	// 前导零宽度：取两边的最大值
	width := len(startStr)
	if len(endStr) > width {
		width = len(endStr)
	}

	// 判断是否有前导零
	hasLeadingZero := false
	if len(startStr) > 1 && startStr[0] == '0' {
		hasLeadingZero = true
	}
	if len(endStr) > 1 && endStr[0] == '0' {
		hasLeadingZero = true
	}

	var result []string
	for i := start; i <= end; i++ {
		if hasLeadingZero {
			result = append(result, fmt.Sprintf("%0*d", width, i))
		} else {
			result = append(result, strconv.Itoa(i))
		}
	}
	return result, nil
}

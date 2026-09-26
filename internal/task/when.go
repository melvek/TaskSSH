package task

import (
	"fmt"
	"strconv"
	"strings"

	"mestrap.com/taskssh/internal/resolve"
)

// evalWhen 求值 when 表达式。
//
// 支持的语法：
//
//	{{var}} == value
//	{{var}} != value
//	{{var}} >  value
//	{{var}} >= value
//	{{var}} <  value
//	{{var}} <= value
//	{{var}} in (a,b,c)
//	expr1 && expr2
//	expr1 || expr2
//
// 求值前先对表达式做变量替换。
// 不支持括号改变优先级；&& 优先级高于 ||。
// 不支持算术运算、函数调用。
func evalWhen(expr string, vars resolve.Vars) (bool, error) {
	if expr == "" {
		return true, nil
	}

	expanded, err := vars.Replace(expr)
	if err != nil {
		return false, err
	}

	// 按 || 分割，任一为 true 则返回 true
	orParts := strings.Split(expanded, "||")
	for _, orPart := range orParts {
		ok, err := evalAnd(orPart)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// evalAnd 处理 && 连接的条件。
func evalAnd(expr string) (bool, error) {
	andParts := strings.Split(expr, "&&")
	for _, part := range andParts {
		ok, err := evalSingle(part)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// evalSingle 处理单个比较。
//
// 注意判断顺序：>= 必须先于 >，<= 必须先于 <，
// 否则会被截断成错误的表达式。
func evalSingle(expr string) (bool, error) {
	expr = strings.TrimSpace(expr)

	switch {
	case strings.Contains(expr, "=="):
		parts := strings.SplitN(expr, "==", 2)
		return trim(parts[0]) == trim(parts[1]), nil

	case strings.Contains(expr, "!="):
		parts := strings.SplitN(expr, "!=", 2)
		return trim(parts[0]) != trim(parts[1]), nil

	case strings.Contains(expr, ">="):
		parts := strings.SplitN(expr, ">=", 2)
		return compare(parts[0], parts[1]) >= 0, nil

	case strings.Contains(expr, "<="):
		parts := strings.SplitN(expr, "<=", 2)
		return compare(parts[0], parts[1]) <= 0, nil

	case strings.Contains(expr, ">"):
		parts := strings.SplitN(expr, ">", 2)
		return compare(parts[0], parts[1]) > 0, nil

	case strings.Contains(expr, "<"):
		parts := strings.SplitN(expr, "<", 2)
		return compare(parts[0], parts[1]) < 0, nil

	case strings.Contains(expr, " in "):
		parts := strings.SplitN(expr, " in ", 2)
		return evalIn(parts[0], parts[1]), nil

	default:
		return false, fmt.Errorf(
			"unsupported when expression: %q "+
				"(supported: == != > >= < <= in, combined with && / ||)",
			expr)
	}
}

// compare 比较两个值。
//
// 两边都能解析为数字时按数字比较，否则按字符串比较。
// 返回 -1、0、1。
func compare(a, b string) int {
	af, errA := strconv.ParseFloat(trim(a), 64)
	bf, errB := strconv.ParseFloat(trim(b), 64)

	if errA == nil && errB == nil {
		switch {
		case af < bf:
			return -1
		case af > bf:
			return 1
		default:
			return 0
		}
	}

	return strings.Compare(trim(a), trim(b))
}

// evalIn 判断 lhs 是否在 rhs 列表中。
//
// rhs 形如 (a,b,c) 或 a,b,c。
func evalIn(lhs, rhs string) bool {
	value := trim(lhs)
	rhs = trim(rhs)

	if strings.HasPrefix(rhs, "(") && strings.HasSuffix(rhs, ")") {
		rhs = rhs[1 : len(rhs)-1]
	}

	for _, item := range strings.Split(rhs, ",") {
		if trim(item) == value {
			return true
		}
	}
	return false
}

// trim 去掉首尾空格。
func trim(s string) string {
	return strings.TrimSpace(s)
}

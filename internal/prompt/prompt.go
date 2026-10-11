// Package prompt 提供交互式输入。
//
// Hidden 用于机密输入（无回显），Line 用于普通输入（有回显）。
// 终端不支持无回显时自动回退到普通读取。
package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Hidden 从终端读取一行输入，无回显。
//
// 终端不支持时回退到普通读取。
func Hidden(label string) (string, error) {
	fmt.Fprint(os.Stderr, label)

	if term.IsTerminal(int(os.Stdin.Fd())) {
		bs, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", fmt.Errorf("read input: %w", err)
		}
		return strings.TrimSpace(string(bs)), nil
	}

	// 回退：普通读取
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	return strings.TrimSpace(line), nil
}

// HiddenConfirm 读取输入并要求二次确认。
//
// 两次不一致时返回错误。
func HiddenConfirm(label string) (string, error) {
	first, err := Hidden(label)
	if err != nil {
		return "", err
	}

	second, err := Hidden("Confirm: ")
	if err != nil {
		return "", err
	}

	if first != second {
		return "", fmt.Errorf("input does not match")
	}

	return first, nil
}

// Line 从终端读取一行输入，有回显。
func Line(label string) (string, error) {
	fmt.Fprint(os.Stderr, label)

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read input: %w", err)
	}
	return strings.TrimSpace(line), nil
}

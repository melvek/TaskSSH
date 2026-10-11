package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/prompt"
	"mestrap.com/taskssh/internal/secret"
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt [text]",
	Short: "Encrypt a string",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text, err := readSecretInput(args, "Text to encrypt: ")
		if err != nil {
			return err
		}

		// 加密需要密钥二次确认，仅在无其它密钥来源时提示
		if !secret.HasKeySource() {
			key, err := secret.PromptKeyConfirm()
			if err != nil {
				return err
			}
			secret.SetRawKey(key)
		}

		cipher, err := secret.Encrypt(text)
		if err != nil {
			return err
		}
		fmt.Println(cipher)
		return nil
	},
}

var decryptCmd = &cobra.Command{
	Use:   "decrypt [cipher]",
	Short: "Decrypt a string",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cipherText, err := readPlainInput(args, "Ciphertext: ")
		if err != nil {
			return err
		}

		plain, err := secret.Decrypt(cipherText)
		if err != nil {
			return err
		}
		fmt.Println(plain)
		return nil
	},
}

// readSecretInput 读取机密文本：优先参数（警告），否则交互输入（隐藏回显 + 二次确认）。
func readSecretInput(args []string, label string) (string, error) {
	if len(args) > 0 {
		warnInsecureArg()
		return args[0], nil
	}
	return prompt.HiddenConfirm(label)
}

// readPlainInput 读取非机密文本：优先参数（警告），否则交互输入（有回显）。
func readPlainInput(args []string, label string) (string, error) {
	if len(args) > 0 {
		warnInsecureArg()
		return args[0], nil
	}
	return prompt.Line(label)
}

// warnInsecureArg 输出命令行传参的安全警告。
func warnInsecureArg() {
	console.Warn("Passing input via command line is insecure and may be recorded in shell history; use interactive input instead")
}

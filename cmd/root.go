package cmd

import (
	"github.com/spf13/cobra"

	"mestrap.com/taskssh/internal/utils"
)

var (
	inventoryFile string
	listOnly      bool
	yesFlag       bool
	portFlag      int
	userFlag      string
	passwordFlag  string
	concurrency   int
)

var rootCmd = &cobra.Command{
	Use:     utils.CLIName + " <task> [hosts...] [options]",
	Short:   utils.ProjectName + " - 轻量级批量运维工具",
	Long:    buildLongDescription(),
	Version: utils.Version,
	Args:    cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return nil
		}
		return runDynamicTask(cmd, args)
	},
}

// Execute 是 CLI 入口。
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 隐藏 completion
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// 全局 flag
	rootCmd.PersistentFlags().StringVarP(&inventoryFile, "inventory", "i",
		utils.DefaultInventory, "Inventory file")
	rootCmd.PersistentFlags().BoolVarP(&listOnly, "list", "l",
		false, "List target hosts only")
	rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y",
		false, "Skip confirmation")
	rootCmd.PersistentFlags().IntVarP(&portFlag, "port", "P",
		0, "Override port")
	rootCmd.PersistentFlags().StringVarP(&userFlag, "user", "u",
		"", "Override username")
	rootCmd.PersistentFlags().StringVarP(&passwordFlag, "password", "p",
		"", "Override password (not recommended)")
	rootCmd.PersistentFlags().IntVarP(&concurrency, "concurrency", "c",
		1, "Number of concurrent connections")

	// 子命令
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(decryptCmd)
	rootCmd.AddCommand(commandCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(fetchCmd)
}

func buildLongDescription() string {
	return utils.ProjectName + ` 是一个基于 SSH 的轻量级批量运维工具。	
https://taskssh.mestrap.com/`
}

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
)

var rootCmd = &cobra.Command{
	Use:     utils.CLIName + " <task> [hosts...] [options]",
	Short:   utils.ProjectName + " - 轻量级批量运维工具",
	Long:    buildLongDescription(),
	Version: utils.Version,
	Args:    cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return runDynamicTask(cmd, args)
	},
}

// Execute 是 CLI 入口。
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(utils.ProjectName + " version {{.Version}}\nBuild time: " + utils.BuildTime)
	// 禁用字母排序，按添加顺序显示
	cobra.EnableCommandSorting = false
	// 隐藏 completion 命令
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

	// 内置子命令
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

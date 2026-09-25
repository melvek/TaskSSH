package cmd

import (
	"github.com/spf13/cobra"

	"mestrap.com/taskssh/internal/secret"
	"mestrap.com/taskssh/internal/utils"
)

var (
	inventoryFile  string
	listOnly       bool
	yesFlag        bool
	portFlag       int
	userFlag       string
	passwordFlag   string
	concurrency    int
	connectTimeout int
	secretKeyFile  string
)

var rootCmd = &cobra.Command{
	Use:           utils.CLIName + " <task> [hosts...] [options]",
	Short:         utils.ProjectName + " - 轻量级批量运维工具",
	Long:          buildLongDescription(),
	Version:       utils.Version,
	Args:          cobra.ArbitraryArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
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
	cobra.EnableCommandSorting = false
	return rootCmd.Execute()
}

func init() {

	rootCmd.SetVersionTemplate(utils.ProjectName + " version {{.Version}}\n")

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

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
	rootCmd.PersistentFlags().IntVarP(&connectTimeout, "connect-timeout", "",
		10, "Connection timeout in seconds")
	rootCmd.PersistentFlags().StringVarP(&secretKeyFile, "secret-key-file", "V",
		"", "File or executable that holds the secret key")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		secret.SecretKeyFile = secretKeyFile
	}

	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(decryptCmd)
	rootCmd.AddCommand(commandCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(scriptCmd)
}

func buildLongDescription() string {
	return utils.ProjectName + ` 是一个基于 SSH 的轻量级批量运维工具。
https://taskssh.mestrap.com/`
}

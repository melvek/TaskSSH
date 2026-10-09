package cmd

import (
	"github.com/spf13/cobra"

	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/secret"
	"mestrap.com/taskssh/internal/utils"
)

var (
	inventoryFile     string
	listOnly          bool
	yesFlag           bool
	portFlag          int
	userFlag          string
	passwordFlag      string
	identityFileFlag  string
	concurrency       int
	connectTimeout    int
	secretKeyFile     string
	dryRun            bool
	cliVars           []string
	serialFlag        int
	minAvailableFlag  int
	stopOnFailureFlag bool
	// execId         string
	verboseFlag bool
	quietFlag   bool
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
		return runDynamicTask(args)
	},
}

// Execute 是 CLI 入口。
func Execute() error {
	cobra.EnableCommandSorting = false
	return rootCmd.Execute()
}

func init() {
	// 覆盖默认的版本号样式
	rootCmd.SetVersionTemplate(utils.ProjectName + " version {{.Version}}\n")
	// 隐藏自动补全命令
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	// 注册全局参数
	rootCmd.PersistentFlags().StringVarP(&inventoryFile, "inventory", "i", utils.DefaultInventory, "Inventory file")
	rootCmd.PersistentFlags().BoolVarP(&listOnly, "list", "l", false, "List target hosts only")
	// rootCmd.PersistentFlags().StringVarP(&execId, "execute-id", "E", "", "Execute ID")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "", false, "Show what would be executed without actually running")
	rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "Skip confirmation")
	rootCmd.PersistentFlags().IntVarP(&portFlag, "port", "P", 0, "port")
	rootCmd.PersistentFlags().StringVarP(&userFlag, "user", "u", "", "username")
	rootCmd.PersistentFlags().StringVarP(&passwordFlag, "password", "p", "", "password (plaintext, not recommended)")
	rootCmd.PersistentFlags().StringVarP(&identityFileFlag, "identity-file", "k", "", "SSH private key file")
	rootCmd.PersistentFlags().IntVarP(&concurrency, "concurrency", "c", 1, "Number of concurrent connections")
	rootCmd.PersistentFlags().IntVarP(&connectTimeout, "connect-timeout", "", 10, "Connection timeout in seconds")
	rootCmd.PersistentFlags().StringVarP(&secretKeyFile, "secret-key-file", "V", "", "File or executable that holds the secret key")
	rootCmd.PersistentFlags().StringArrayVarP(&cliVars, "define", "D", nil, "Set variable (key=value), can be repeated")
	rootCmd.PersistentFlags().IntVarP(&serialFlag, "serial", "S", -1, "Run in batches of N hosts per group (0 = all at once; default: use task config)")
	rootCmd.PersistentFlags().IntVarP(&minAvailableFlag, "min-available", "", -1, "Keep at least N hosts per group running (0 = no limit; default: use task config)")
	rootCmd.PersistentFlags().BoolVarP(&stopOnFailureFlag, "stop-on-failure", "", true, "Stop remaining batches if a batch fails")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Only show errors")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		secret.SecretKeyFile = secretKeyFile

		switch {
		case dryRun || verboseFlag:
			console.SetLevel(console.LevelVerbose)
		case quietFlag:
			console.SetLevel(console.LevelQuiet)
		default:
			console.SetLevel(console.LevelBrief)
		}

		if passwordFlag != "" {
			console.Warn("Passing password via command line is insecure; use 'taskssh encrypt' and inventory file instead")
		}
	}

	// 注册子命令
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(decryptCmd)
	rootCmd.AddCommand(commandCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(scriptCmd)
}

func buildLongDescription() string {
	return utils.ProjectName + " 是一个基于 SSH 的轻量级批量运维工具。\nhttps://taskssh.mestrap.com/"
}

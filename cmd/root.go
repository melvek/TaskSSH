package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/fileutil"
	"mestrap.com/taskssh/internal/secret"
)

var (
	inventoryFile     string
	listOnly          bool
	confirmFlag       bool
	targetHostPattern string
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
	verboseFlag       bool
	quietFlag         bool
	traceFlag         bool
)

var rootCmd = &cobra.Command{
	Use:           "taskssh <task> <host-pattern> [options]",
	Short:         "TaskSSH is a lightweight operations tool.",
	Long:          "TaskSSH is a lightweight, SSH-based batch operations tool.\nhttps://taskssh.mestrap.com/",
	Version:       Version,
	Args:          cobra.ArbitraryArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return nil
		}
		return runDynamicTask(cmd.Context(), args)
	},
}

// Execute 是 CLI 入口。
func Execute(ctx context.Context) error {
	cobra.EnableCommandSorting = false
	return rootCmd.ExecuteContext(ctx)
}

func init() {

	// 覆盖默认的版本号样式
	rootCmd.SetVersionTemplate("TaskSSH version {{.Version}}\n")

	// 隐藏自动补全命令
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// 注册全局参数
	rootCmd.PersistentFlags().StringVarP(&inventoryFile, "inventory", "i", "", "Inventory file (default: inventory.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&listOnly, "list", "l", false, "List target hosts only")
	rootCmd.PersistentFlags().StringArrayVarP(&cliVars, "define", "D", nil, "Set variable (key=value), can be repeated")

	// 服务器信息
	rootCmd.PersistentFlags().IntVarP(&portFlag, "port", "P", 0, "port")
	rootCmd.PersistentFlags().StringVarP(&userFlag, "user", "u", "", "username")
	rootCmd.PersistentFlags().StringVarP(&passwordFlag, "password", "p", "", "password (plaintext, not recommended)")

	// 加密及登录
	rootCmd.PersistentFlags().StringVarP(&identityFileFlag, "identity-file", "k", "", "SSH private key file")
	rootCmd.PersistentFlags().StringVarP(&secretKeyFile, "secret-key-file", "V", "", "File or executable that holds the secret key")

	// 执行控制
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "", false, "Show what would be executed without actually running")
	rootCmd.PersistentFlags().BoolVarP(&confirmFlag, "confirm", "", false, "List target hosts and wait for confirmation before executing")
	rootCmd.PersistentFlags().IntVarP(&concurrency, "concurrency", "c", 1, "Number of concurrent connections")
	rootCmd.PersistentFlags().IntVarP(&connectTimeout, "connect-timeout", "", 10, "Connection timeout in seconds")
	rootCmd.PersistentFlags().IntVarP(&serialFlag, "serial", "S", -1, "Run in batches of N hosts per group (0 = all at once; default: use task config)")
	rootCmd.PersistentFlags().IntVarP(&minAvailableFlag, "min-available", "", -1, "Keep at least N hosts per group running (0 = no limit; default: use task config)")
	rootCmd.PersistentFlags().BoolVarP(&stopOnFailureFlag, "stop-on-failure", "", true, "Stop remaining batches if a batch fails")

	// 日志输出相关
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Only show errors")
	rootCmd.PersistentFlags().BoolVar(&traceFlag, "trace", false, "Enable trace output")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		secret.SecretKeyFile = secretKeyFile

		if inventoryFile != "" && !fileutil.FileExists(inventoryFile) {
			return fmt.Errorf("inventory file not found: %s", inventoryFile)
		}

		if secretKeyFile != "" && !fileutil.FileExists(secretKeyFile) {
			return fmt.Errorf("secret key file not found: %s", secretKeyFile)
		}

		if identityFileFlag != "" && !fileutil.FileExists(identityFileFlag) {
			return fmt.Errorf("secret key file not found: %s", identityFileFlag)
		}

		switch {
		case traceFlag:
			console.SetLevel(console.LevelTrace)
		case dryRun || verboseFlag:
			console.SetLevel(console.LevelDebug)
		case quietFlag:
			console.SetLevel(console.LevelError)
		default:
			console.SetLevel(console.LevelInfo)
		}

		if passwordFlag != "" {
			console.Default.Warn("Passing password via command line is insecure; use 'taskssh encrypt' and inventory file instead")
		}

		return nil
	}

	// 注册子命令
	rootCmd.AddCommand(commandCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(fetchCmd)
	rootCmd.AddCommand(scriptCmd)
	rootCmd.AddCommand(keygenCmd)
	rootCmd.AddCommand(encryptCmd)
	rootCmd.AddCommand(decryptCmd)
}

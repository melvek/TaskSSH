package cmd

import (
	"net"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"mestrap.com/taskssh/internal/config"
	"mestrap.com/taskssh/internal/console"
	"mestrap.com/taskssh/internal/host"
	"mestrap.com/taskssh/internal/ssh"
)

var testCmd = &cobra.Command{
	Use:   "test <host-pattern>",
	Short: "Test inventory configuration and SSH connections",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inv, err := config.Load(inventoryFile)
		if err != nil {
			return err
		}

		entries, err := host.ResolveHosts(args[0], inv, host.Overrides{
			Port:         portFlag,
			Username:     userFlag,
			Password:     passwordFlag,
			IdentityFile: identityFileFlag,
		})
		if err != nil {
			return err
		}

		ssh.ClearPasswordCache()
		defer ssh.ClearPasswordCache()

		for _, entry := range entries {
			port := entry.Host.Port
			if port == 0 {
				port = ssh.DefaultPort
			}
			address := net.JoinHostPort(entry.Host.Host, strconv.Itoa(port))
			console.Default.Debug("Testing SSH connection to %s (%s)", entry.Name, address)

			client, err := ssh.ConnectContext(cmd.Context(), &entry.Host, time.Duration(connectTimeout)*time.Second)
			if err != nil {
				console.Default.Error("SSH connection failed for %s: %v", entry.Name, err)
				continue
			}
			if err := client.Close(); err != nil {
				console.Default.Error("SSH connection cleanup failed for %s: %v", entry.Name, err)
				continue
			}
			console.Default.Info("SSH connection successful: %s (%s)", entry.Name, address)
		}

		return nil
	},
}

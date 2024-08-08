package cmd

import (
	"fmt"

	"github.com/fredytarigan/Tendo/pkg/tendo/logger"
	"github.com/spf13/cobra"
)

type CommandEngine struct {
	rootCmd *cobra.Command
}

func NewCommandEngine() *CommandEngine {
	var rootCmd = &cobra.Command {
		Use: "tendo",
		Short: "tendo CLI",
		Long: "tendo service command line",
	}

	return &CommandEngine{
		rootCmd: rootCmd,
	}
}

func (c *CommandEngine) Run() {
	defer logger.Logger.Sync()

	var kubeconfig string

	var commands = []*cobra.Command {
		{
			Use: "server",
			Short: "tendo start HTTP server",
			Long: "command to start HTTP server of tendo service",
			Run: func(cmd *cobra.Command, args []string) {
				kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
				
				ServerListen(kubeconfig)
			},
		},
	}

	for _, command := range commands {
		c.rootCmd.AddCommand(command)

		if command.Name() == "server" {
			c.rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "add kubeconfig file path")
		}
	}

	if err := c.rootCmd.Execute(); err != nil {
		msg := fmt.Sprintf("failed to execute command with error: %s", err)
		logger.Logger.Error(msg)
	}
}
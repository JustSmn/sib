package cmd

/*
root.go — это ядро CLI
Он содержит:

Корневую команду (sib)
Регистрацию всех подкоманд (init, add, commit)
Глобальные настройки (флаги, версия, help)
*/

import (
	"fmt"
	"os"
	"sib/internal/cli"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sib",
	Short: "Sib - a Git-like version control system",
	Long: `Sib is a distributed version control system similar to Git.
It helps you track changes in your source code during software development.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Если команда запущена без аргументов
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Глобальные флаги (если понадобятся)
	rootCmd.AddCommand(cli.InitCmd)
	rootCmd.AddCommand(cli.AddCmd)
	//rootCmd.AddCommand(cli.CommitCmd)
}

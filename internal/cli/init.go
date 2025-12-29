package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"sib/internal/commands"
)

// InitCmd - cobra команда для init
var InitCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new Sib repository",
	Long: `Initialize a new, empty Sib repository in the specified directory.
If no directory is provided, uses the current directory.`,
	Args: cobra.MaximumNArgs(1), // максимум 1 аргумент
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		if err := commands.Init(path); err != nil {
			fmt.Printf("error: %v\n", err)
		}
	},
}

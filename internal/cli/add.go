// internal/cli/add.go
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"sib/internal/commands"
)

// AddCmd - cobra команда для add
var AddCmd = &cobra.Command{
	Use:   "add [path...]",
	Short: "Add file contents to the index",
	Long: `Add files to the staging area (index). 
Use '.' to add all files in the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			args = []string{"."}
		}

		// Пока поддерживаем только sib add .
		if len(args) == 1 && args[0] == "." {
			if err := commands.Add("."); err != nil {
				fmt.Printf("error: %v\n", err)
			}
		} else {
			fmt.Println("error: only 'sib add .' is supported for now")
		}
	},
}

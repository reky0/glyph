package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Print raw text of an entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, _, err := loadEntries()
		if err != nil {
			return err
		}
		entry, _, err := findByID(entries, args[0])
		if err != nil {
			return err
		}
		fmt.Print(entry.Text)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}

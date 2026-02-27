package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm <id>",
	Short: "Remove an entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entries, s, err := loadEntries()
		if err != nil {
			return err
		}
		_, idx, err := findByID(entries, args[0])
		if err != nil {
			return err
		}
		updated := append(entries[:idx], entries[idx+1:]...)
		if err := s.Save(updated); err != nil {
			return err
		}
		fmt.Printf("removed %s\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
}

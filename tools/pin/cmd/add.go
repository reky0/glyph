package cmd

import (
	"fmt"
	"strings"

	store "github.com/reky0/glyph-store"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <text>",
	Short: "Save a new entry",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		tag, _ := cmd.Flags().GetString("tag")
		isURL, _ := cmd.Flags().GetBool("url")
		isCmd, _ := cmd.Flags().GetBool("cmd")

		entryType := ""
		switch {
		case isURL:
			entryType = "url"
		case isCmd:
			entryType = "cmd"
		default:
			entryType = InferType(text)
		}

		entry := PinEntry{
			Entry: store.NewEntry(),
			Text:  text,
			Tag:   tag,
			Type:  entryType,
		}

		s, err := openStore()
		if err != nil {
			return err
		}
		if err := s.Append(entry); err != nil {
			return err
		}

		fmt.Printf("pinned %s [%s]\n", shortID(entry.ID), entryType)
		return nil
	},
}

func init() {
	addCmd.Flags().String("tag", "", "Tag for the entry")
	addCmd.Flags().Bool("url", false, "Mark entry as a URL")
	addCmd.Flags().Bool("cmd", false, "Mark entry as a command")
	rootCmd.AddCommand(addCmd)
}

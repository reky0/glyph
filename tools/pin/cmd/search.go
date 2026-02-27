package cmd

import (
	"strings"
	"time"

	ink "github.com/reky0/glyph-ink"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search entries by text or tag",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.ToLower(args[0])
		entries, _, err := loadEntries()
		if err != nil {
			return err
		}

		theme := ink.ThemeFrom(viper.GetString("style"))
		tbl := theme.Table().Headers("ID", "TYPE", "TAG", "TEXT", "DATE")

		for _, e := range entries {
			if strings.Contains(strings.ToLower(e.Text), query) ||
				strings.Contains(strings.ToLower(e.Tag), query) {
				tbl.Row(
					shortID(e.ID),
					e.Type,
					e.Tag,
					truncate(e.Text, 60),
					e.CreatedAt.Format(time.DateOnly),
				)
			}
		}

		tbl.RenderToStdout()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

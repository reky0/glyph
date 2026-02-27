package cmd

import (
	"time"

	ink "github.com/reky0/glyph-ink"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List pinned entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		filterTag, _ := cmd.Flags().GetString("tag")
		filterType, _ := cmd.Flags().GetString("type")

		entries, _, err := loadEntries()
		if err != nil {
			return err
		}

		theme := ink.ThemeFrom(viper.GetString("style"))
		tbl := theme.Table().Headers("ID", "TYPE", "TAG", "TEXT", "DATE")

		for _, e := range entries {
			if filterTag != "" && e.Tag != filterTag {
				continue
			}
			if filterType != "" && e.Type != filterType {
				continue
			}
			tbl.Row(
				shortID(e.ID),
				e.Type,
				e.Tag,
				truncate(e.Text, 60),
				e.CreatedAt.Format(time.DateOnly),
			)
		}

		tbl.RenderToStdout()
		return nil
	},
}

func init() {
	listCmd.Flags().String("tag", "", "Filter by tag")
	listCmd.Flags().String("type", "", "Filter by type: url, cmd, note")
	rootCmd.AddCommand(listCmd)
}

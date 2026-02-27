package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is injected at build time via ldflags.
var Version = "dev"

var rootCmd = &cobra.Command{
	Use:     "ask <question>",
	Short:   "Ask a question to an AI with automatic directory context",
	Version: Version,
	Args:    cobra.MinimumNArgs(1),
	RunE:    runAsk,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("style", "rounded", "Output style: ascii, rounded, minimal")
	rootCmd.Flags().Bool("no-context", false, "Skip automatic directory context injection")
	if err := viper.BindPFlag("style", rootCmd.PersistentFlags().Lookup("style")); err != nil {
		panic(fmt.Sprintf("failed to bind style flag: %v", err))
	}
}

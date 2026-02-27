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
	Use:     "pin",
	Short:   "Clipboard for things you find in the terminal",
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("style", "rounded", "Output style: ascii, rounded, minimal")
	if err := viper.BindPFlag("style", rootCmd.PersistentFlags().Lookup("style")); err != nil {
		panic(fmt.Sprintf("failed to bind style flag: %v", err))
	}
}

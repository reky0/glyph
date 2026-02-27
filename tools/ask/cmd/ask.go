package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	core "github.com/reky0/glyph-core"
	ink "github.com/reky0/glyph-ink"
	mind "github.com/reky0/glyph-mind"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const systemPromptTmpl = `You are a helpful terminal assistant. Be concise. Use plain text, no markdown headers.
When relevant, prefer showing commands over explaining them.`

func runAsk(cmd *cobra.Command, args []string) error {
	question := strings.Join(args, " ")
	noContext, _ := cmd.Flags().GetBool("no-context")

	// Read piped stdin if available.
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		piped, err := io.ReadAll(os.Stdin)
		if err == nil && len(piped) > 0 {
			question = string(piped) + "\n\n" + question
		}
	}

	cfg, err := core.LoadConfig()
	if err != nil {
		return err
	}
	if style := viper.GetString("style"); style != "" {
		cfg.DefaultStyle = style
	}

	client, err := mind.NewClientFromConfig(cfg)
	if err != nil {
		theme := ink.ThemeFrom(cfg.DefaultStyle)
		fmt.Fprintln(os.Stderr, theme.Error(err.Error()))
		os.Exit(1)
	}

	systemPrompt := systemPromptTmpl
	if !noContext {
		cwd, err := os.Getwd()
		if err == nil {
			ctx := gatherContext(cwd)
			if ctx != "" {
				systemPrompt += "\n\nCurrent directory context:\n" + ctx
			}
		}
	}

	ch, err := client.Stream(context.Background(), systemPrompt, question)
	if err != nil {
		theme := ink.ThemeFrom(cfg.DefaultStyle)
		fmt.Fprintln(os.Stderr, theme.Error(err.Error()))
		os.Exit(1)
	}

	printer := ink.NewStreamPrinter(os.Stdout)
	return printer.PrintStream(ch)
}

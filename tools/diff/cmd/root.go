package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	core "github.com/reky0/glyph-core"
	ink "github.com/reky0/glyph-ink"
	mind "github.com/reky0/glyph-mind"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is injected at build time via ldflags.
var Version = "dev"

const diffSystemPrompt = `You are a code reviewer. Summarize what this diff does in plain language.
List the most important changes as a short bullet list.
End with one line flagging any potential issue if you see one, or "Looks clean." if not.`

var rootCmd = &cobra.Command{
	Use:     "diff",
	Short:   "Explain a git diff using AI",
	Version: Version,
	RunE:    runDiff,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("style", "rounded", "Output style: ascii, rounded, minimal")
	rootCmd.Flags().Bool("staged", false, "Diff staged changes (git diff --cached)")
	rootCmd.Flags().String("commit", "", "Explain a specific commit (git show <hash>)")
	if err := viper.BindPFlag("style", rootCmd.PersistentFlags().Lookup("style")); err != nil {
		panic(fmt.Sprintf("failed to bind style flag: %v", err))
	}
}

func runDiff(cmd *cobra.Command, args []string) error {
	theme := ink.ThemeFrom(viper.GetString("style"))

	staged, _ := cmd.Flags().GetBool("staged")
	commitHash, _ := cmd.Flags().GetString("commit")

	diffOutput, err := getDiff(staged, commitHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, theme.Error(err.Error()))
		os.Exit(1)
	}
	if len(bytes.TrimSpace(diffOutput)) == 0 {
		fmt.Println(theme.Muted("No changes found."))
		return nil
	}

	cfg, err := core.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, theme.Error(err.Error()))
		os.Exit(1)
	}
	if style := viper.GetString("style"); style != "" {
		cfg.DefaultStyle = style
	}

	client, err := mind.NewClientFromConfig(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, theme.Error(err.Error()))
		os.Exit(1)
	}

	ch, err := client.Stream(context.Background(), diffSystemPrompt, string(diffOutput))
	if err != nil {
		fmt.Fprintln(os.Stderr, theme.Error(err.Error()))
		os.Exit(1)
	}

	printer := ink.NewStreamPrinter(os.Stdout)
	return printer.PrintStream(ch)
}

func getDiff(staged bool, commitHash string) ([]byte, error) {
	var gitArgs []string
	switch {
	case commitHash != "":
		gitArgs = []string{"show", commitHash}
	case staged:
		gitArgs = []string{"diff", "--cached"}
	default:
		gitArgs = []string{"diff", "HEAD"}
	}

	cmd := exec.Command("git", gitArgs...)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		msg := errBuf.String()
		if msg == "" {
			msg = err.Error()
		}
		// Check for non-git-repo error message.
		return nil, &core.AppError{
			Msg: "git command failed — are you inside a git repository?",
			Err: fmt.Errorf("%s", msg),
		}
	}
	return out.Bytes(), nil
}

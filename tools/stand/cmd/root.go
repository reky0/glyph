package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	core "github.com/reky0/glyph-core"
	ink "github.com/reky0/glyph-ink"
	mind "github.com/reky0/glyph-mind"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is injected at build time via ldflags.
var Version = "dev"

const standSystemPrompt = `You are helping a developer write a standup update.
Given a list of git commit messages, write a short standup in first person.
Format: 3-5 bullet points, plain English, no jargon, no markdown.
Focus on what was done, not implementation details.`

var rootCmd = &cobra.Command{
	Use:     "stand",
	Short:   "Generate a standup update from recent git activity",
	Version: Version,
	RunE:    runStand,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("style", "rounded", "Output style: ascii, rounded, minimal")
	rootCmd.Flags().String("since", "today", "Date range: today, yesterday, '2 days ago', or any git-compatible date")
	rootCmd.Flags().Bool("copy", false, "Print a note to pipe output manually to clipboard")
	if err := viper.BindPFlag("style", rootCmd.PersistentFlags().Lookup("style")); err != nil {
		panic(fmt.Sprintf("failed to bind style flag: %v", err))
	}
}

func runStand(cmd *cobra.Command, args []string) error {
	theme := ink.ThemeFrom(viper.GetString("style"))
	since, _ := cmd.Flags().GetString("since")
	copyMode, _ := cmd.Flags().GetBool("copy")

	commits, err := getCommits(since)
	if err != nil {
		fmt.Fprintln(os.Stderr, theme.Error(err.Error()))
		os.Exit(1)
	}
	if strings.TrimSpace(commits) == "" {
		fmt.Println(theme.Muted("No commits found since " + since + "."))
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

	ch, err := client.Stream(context.Background(), standSystemPrompt, commits)
	if err != nil {
		fmt.Fprintln(os.Stderr, theme.Error(err.Error()))
		os.Exit(1)
	}

	printer := ink.NewStreamPrinter(os.Stdout)
	if err := printer.PrintStream(ch); err != nil {
		return err
	}

	if copyMode {
		fmt.Fprintln(os.Stderr, theme.Muted("\nTip: pipe output to clipboard with: stand | pbcopy  (macOS) or  stand | xclip  (Linux)"))
	}
	return nil
}

func getCommits(since string) (string, error) {
	// Resolve "today" / "yesterday" to git-compatible values.
	switch strings.ToLower(strings.TrimSpace(since)) {
	case "today":
		since = "midnight"
	case "yesterday":
		since = "yesterday midnight"
	}

	// Get the author email from git config.
	authorBytes, err := runGit("config", "user.email")
	if err != nil {
		// Proceed without --author filter if git config fails.
		authorBytes = nil
	}
	author := strings.TrimSpace(string(authorBytes))

	gitArgs := []string{
		"log",
		"--since=" + since,
		"--pretty=format:%s",
	}
	if author != "" {
		gitArgs = append(gitArgs, "--author="+author)
	}

	out, err := runGit(gitArgs...)
	if err != nil {
		return "", &core.AppError{
			Msg: "git log failed — are you inside a git repository?",
			Err: err,
		}
	}
	return string(out), nil
}

func runGit(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		msg := errBuf.String()
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return out.Bytes(), nil
}

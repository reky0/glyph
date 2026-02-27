package cmd

import (
	"net/url"
	"strings"

	store "github.com/reky0/glyph-store"
)

// PinEntry is a single stored item.
type PinEntry struct {
	store.Entry
	Text string `json:"text"`
	Tag  string `json:"tag"`
	Type string `json:"type"` // url | cmd | note
}

// InferType guesses the entry type from the text content.
func InferType(text string) string {
	if looksLikeURL(text) {
		return "url"
	}
	if looksLikeCmd(text) {
		return "cmd"
	}
	return "note"
}

func looksLikeURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https" || u.Scheme == "ftp")
}

func looksLikeCmd(s string) bool {
	cmdPrefixes := []string{
		"sudo ", "git ", "go ", "npm ", "docker ", "kubectl ",
		"make ", "ls ", "cd ", "cat ", "grep ", "awk ", "sed ",
		"curl ", "wget ", "ssh ", "scp ",
	}
	lower := strings.ToLower(strings.TrimSpace(s))
	for _, p := range cmdPrefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

// shortID returns the first 8 characters of the entry ID.
func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// truncate shortens s to n runes, appending "…" if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

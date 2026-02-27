package ink

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Theme defines the visual contract for all glyph output.
type Theme interface {
	// Header renders a section header string.
	Header(s string) string
	// Muted renders de-emphasized text.
	Muted(s string) string
	// Success renders a success message.
	Success(s string) string
	// Error renders an error message.
	Error(s string) string
	// Table returns a pre-styled table renderer.
	Table() *TableRenderer
}

// TableRenderer holds column headers and rows and can print a styled table.
type TableRenderer struct {
	headers []string
	rows    [][]string
	style   tableStyle
}

type tableStyle int

const (
	tableASCII   tableStyle = iota
	tableRounded tableStyle = iota
	tableMinimal tableStyle = iota
)

func newTable(s tableStyle) *TableRenderer {
	return &TableRenderer{style: s}
}

func (t *TableRenderer) Headers(h ...string) *TableRenderer {
	t.headers = h
	return t
}

func (t *TableRenderer) Row(cols ...string) *TableRenderer {
	t.rows = append(t.rows, cols)
	return t
}

func (t *TableRenderer) Render(w io.Writer) {
	if len(t.headers) == 0 {
		return
	}

	// Compute column widths.
	widths := make([]int, len(t.headers))
	for i, h := range t.headers {
		widths[i] = len(h)
	}
	for _, row := range t.rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	accent := lipgloss.Color("#7C6AF7")
	muted := lipgloss.Color("#6C6C6C")

	switch t.style {
	case tableASCII:
		t.renderASCII(w, widths, muted)
	case tableRounded:
		t.renderRounded(w, widths, accent)
	case tableMinimal:
		t.renderMinimal(w, widths, muted)
	}
}

func pad(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

func (t *TableRenderer) renderASCII(w io.Writer, widths []int, muted lipgloss.Color) {
	headerStyle := lipgloss.NewStyle().Foreground(muted).Bold(true)
	sep := "+"
	for _, w := range widths {
		sep += strings.Repeat("-", w+2) + "+"
	}
	fmt.Fprintln(w, sep)
	row := "|"
	for i, h := range t.headers {
		row += " " + headerStyle.Render(pad(h, widths[i])) + " |"
	}
	fmt.Fprintln(w, row)
	fmt.Fprintln(w, sep)
	for _, r := range t.rows {
		line := "|"
		for i, cell := range r {
			if i < len(widths) {
				line += " " + pad(cell, widths[i]) + " |"
			}
		}
		fmt.Fprintln(w, line)
	}
	fmt.Fprintln(w, sep)
}

func (t *TableRenderer) renderRounded(w io.Writer, widths []int, accent lipgloss.Color) {
	headerStyle := lipgloss.NewStyle().Foreground(accent).Bold(true)

	// Build border pieces manually for rounded look.
	totalWidth := 0
	for _, ww := range widths {
		totalWidth += ww + 3
	}
	totalWidth++

	top := "\u256d" + strings.Repeat("\u2500", totalWidth-2) + "\u256e"
	mid := "\u251c" + strings.Repeat("\u2500", totalWidth-2) + "\u2524"
	bot := "\u2570" + strings.Repeat("\u2500", totalWidth-2) + "\u256f"

	fmt.Fprintln(w, top)
	row := "\u2502"
	for i, h := range t.headers {
		row += " " + headerStyle.Render(pad(h, widths[i])) + " \u2502"
	}
	fmt.Fprintln(w, row)
	fmt.Fprintln(w, mid)
	for _, r := range t.rows {
		line := "\u2502"
		for i, cell := range r {
			if i < len(widths) {
				line += " " + pad(cell, widths[i]) + " \u2502"
			}
		}
		fmt.Fprintln(w, line)
	}
	fmt.Fprintln(w, bot)
}

func (t *TableRenderer) renderMinimal(w io.Writer, widths []int, muted lipgloss.Color) {
	headerStyle := lipgloss.NewStyle().Foreground(muted)
	row := ""
	for i, h := range t.headers {
		if i > 0 {
			row += "  "
		}
		row += headerStyle.Render(pad(h, widths[i]))
	}
	fmt.Fprintln(w, row)
	// Underline headers.
	under := ""
	for i, ww := range widths {
		if i > 0 {
			under += "  "
		}
		under += strings.Repeat("-", ww)
	}
	fmt.Fprintln(w, under)
	for _, r := range t.rows {
		line := ""
		for i, cell := range r {
			if i < len(widths) {
				if i > 0 {
					line += "  "
				}
				line += pad(cell, widths[i])
			}
		}
		fmt.Fprintln(w, line)
	}
}

// RenderToStdout is a convenience wrapper around Render(os.Stdout).
func (t *TableRenderer) RenderToStdout() {
	t.Render(os.Stdout)
}

// ─── ASCII theme ─────────────────────────────────────────────────────────────

type asciiTheme struct{}

var _ Theme = asciiTheme{}

func (asciiTheme) Header(s string) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A8A8A8")).
		Bold(true).
		Render("=== " + s + " ===")
}

func (asciiTheme) Muted(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C")).Render(s)
}

func (asciiTheme) Success(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#A8A8A8")).Render("[ok] " + s)
}

func (asciiTheme) Error(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#A8A8A8")).Render("[err] " + s)
}

func (asciiTheme) Table() *TableRenderer { return newTable(tableASCII) }

// ─── Rounded theme ───────────────────────────────────────────────────────────

type roundedTheme struct{}

var _ Theme = roundedTheme{}

var accent = lipgloss.Color("#7C6AF7")

func (roundedTheme) Header(s string) string {
	return lipgloss.NewStyle().
		Foreground(accent).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accent).
		Padding(0, 1).
		Render(s)
}

func (roundedTheme) Muted(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C")).Render(s)
}

func (roundedTheme) Success(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render("✓ " + s)
}

func (roundedTheme) Error(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("✗ " + s)
}

func (roundedTheme) Table() *TableRenderer { return newTable(tableRounded) }

// ─── Minimal theme ───────────────────────────────────────────────────────────

type minimalTheme struct{}

var _ Theme = minimalTheme{}

var minAccent = lipgloss.Color("#A8A8A8")

func (minimalTheme) Header(s string) string {
	return lipgloss.NewStyle().Foreground(minAccent).Bold(true).Render(s)
}

func (minimalTheme) Muted(s string) string {
	return lipgloss.NewStyle().Foreground(minAccent).Render(s)
}

func (minimalTheme) Success(s string) string {
	return lipgloss.NewStyle().Foreground(minAccent).Render("ok  " + s)
}

func (minimalTheme) Error(s string) string {
	return lipgloss.NewStyle().Foreground(minAccent).Render("err " + s)
}

func (minimalTheme) Table() *TableRenderer { return newTable(tableMinimal) }

// ─── Factory ─────────────────────────────────────────────────────────────────

// ThemeFrom returns a Theme for the given name.
// Valid values: "ascii", "rounded", "minimal". Defaults to "rounded".
func ThemeFrom(name string) Theme {
	switch strings.ToLower(name) {
	case "ascii":
		return asciiTheme{}
	case "minimal":
		return minimalTheme{}
	default:
		return roundedTheme{}
	}
}

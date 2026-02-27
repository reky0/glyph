package ink

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// StreamPrinter writes AI-streamed text chunks to an output writer,
// flushing after every chunk so the user sees output in real time.
type StreamPrinter struct {
	w io.Writer
}

// NewStreamPrinter returns a StreamPrinter writing to w.
// Pass os.Stdout for terminal use.
func NewStreamPrinter(w io.Writer) *StreamPrinter {
	return &StreamPrinter{w: w}
}

// Print writes a single chunk to the output, without a trailing newline.
func (p *StreamPrinter) Print(chunk string) error {
	_, err := fmt.Fprint(p.w, chunk)
	return err
}

// PrintStream consumes a channel of string chunks and prints each one.
// It writes a trailing newline when the channel is closed.
func (p *StreamPrinter) PrintStream(ch <-chan string) error {
	bw := bufio.NewWriter(p.w)
	for chunk := range ch {
		if _, err := fmt.Fprint(bw, chunk); err != nil {
			return err
		}
		if err := bw.Flush(); err != nil {
			return err
		}
	}
	// Ensure we end on a new line.
	_, err := fmt.Fprintln(p.w)
	return err
}

// DefaultStreamPrinter is a StreamPrinter writing to os.Stdout.
var DefaultStreamPrinter = NewStreamPrinter(os.Stdout)

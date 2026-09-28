// Package render formats command output: aligned tables, labelled fields
// and JSON. Text read from a stick can hold anything, so every cell is
// sanitised before it reaches the terminal.
package render

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
)

// Sanitise returns s with control characters written as Go escapes, so a
// name cannot break a table or send escape sequences to the terminal.
func Sanitise(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) {
			q := strconv.QuoteRune(r) // for example '\t' or '\x1b'
			b.WriteString(q[1 : len(q)-1])
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Table writes header and rows as columns two spaces apart. Widths come
// from the content, measured in terminal cells so wide characters line up.
// The last column is not padded. Every row must have len(header) cells.
func Table(w io.Writer, header []string, rows [][]string) error {
	all := append([][]string{header}, rows...)
	widths := make([]int, len(header))
	clean := make([][]string, len(all))
	for i, row := range all {
		clean[i] = make([]string, len(row))
		for j, cell := range row {
			c := Sanitise(cell)
			clean[i][j] = c
			widths[j] = max(widths[j], runewidth.StringWidth(c))
		}
	}
	var b strings.Builder
	for _, row := range clean {
		for j, c := range row {
			b.WriteString(c)
			if j < len(row)-1 {
				b.WriteString(strings.Repeat(" ", widths[j]-runewidth.StringWidth(c)+2))
			}
		}
		b.WriteByte('\n')
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// Fields writes title, then one indented "Label:" line per pair with the
// values aligned two spaces after the widest label.
func Fields(w io.Writer, title string, pairs [][2]string) error {
	width := 0
	for _, p := range pairs {
		width = max(width, runewidth.StringWidth(p[0])+1)
	}
	var b strings.Builder
	b.WriteString(Sanitise(title) + "\n")
	for _, p := range pairs {
		label := Sanitise(p[0]) + ":"
		b.WriteString("  " + label + strings.Repeat(" ", width-runewidth.StringWidth(label)+2) + Sanitise(p[1]) + "\n")
	}
	_, err := io.WriteString(w, b.String())
	return err
}

// JSON writes v as indented JSON followed by a newline.
func JSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

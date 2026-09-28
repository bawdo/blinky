package app

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/bawdo/blinky/internal/exitcode"
	"github.com/bawdo/blinky/internal/render"
	"github.com/bawdo/blinky/internal/target"
)

type blockValue struct {
	Text string `json:"text"`
	Hex  string `json:"hex"`
}

type blockJSON struct {
	ID        string     `json:"id"`
	Serial    string     `json:"serial"`
	Block     int        `json:"block"`
	InfoBlock blockValue `json:"info_block"`
}

// ReadInfoBlock prints info block n of each chosen stick as text and hex.
func (a *App) ReadInfoBlock(req target.Request, n int, asJSON bool) error {
	if err := checkBlock(n); err != nil {
		return err
	}
	infos, _, err := a.resolve(req, target.Group)
	if err != nil {
		return err
	}
	blocks := make([][]byte, len(infos))
	runErr := a.each(infos, nil, func(o opened) error {
		b, err := o.st.InfoBlock(n)
		if err != nil {
			return libError(err)
		}
		blocks[o.pos] = append([]byte{}, b...) // non-nil even when empty
		return nil
	})
	rows := make([]blockJSON, 0, len(infos))
	for i, b := range blocks {
		if b != nil {
			rows = append(rows, blockJSON{ID: infos[i].ID(), Serial: infos[i].Serial, Block: n,
				InfoBlock: blockValue{Text: string(b), Hex: spacedHex(b)}})
		}
	}
	if err := printRows(a, asJSON, rows, func(r blockJSON) (string, string) {
		if r.InfoBlock.Hex == "" {
			return r.ID, "-"
		}
		return r.ID, fmt.Sprintf("%s  (%s)", render.Sanitise(r.InfoBlock.Text), r.InfoBlock.Hex)
	}); err != nil {
		return err
	}
	return runErr
}

// SetInfoBlock writes data to info block n of one stick, as text, or as
// bytes when isHex is set. "" clears the block.
func (a *App) SetInfoBlock(req target.Request, n int, data string, isHex bool) error {
	if err := checkBlock(n); err != nil {
		return err
	}
	b := []byte(data)
	if isHex {
		var err error
		if b, err = hex.DecodeString(strings.ReplaceAll(data, " ", "")); err != nil {
			return exitcode.Invalid("%q is not hex bytes, such as 68656c6c6f or \"68 65 6c 6c 6f\"", data)
		}
	}
	if len(b) > maxBlock {
		return exitcode.Invalid("%d bytes, an info block holds at most %d", len(b), maxBlock)
	}
	infos, _, err := a.resolve(req, target.One)
	if err != nil {
		return err
	}
	if n == 1 {
		_, _ = fmt.Fprintln(a.err, "warning: info block 1 holds the stick's name, so this replaces it")
	}
	return a.each(infos, nil, func(o opened) error { return libError(o.st.SetInfoBlock(n, b)) })
}

func checkBlock(n int) error {
	if n != 1 && n != 2 {
		return exitcode.Invalid("info block %d, want 1 or 2", n)
	}
	return nil
}

// spacedHex writes b as "68 65 6c".
func spacedHex(b []byte) string {
	parts := make([]string, len(b))
	for i, c := range b {
		parts[i] = fmt.Sprintf("%02x", c)
	}
	return strings.Join(parts, " ")
}

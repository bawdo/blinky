package effect

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestImportsOnlyStdlibAndBlinkstick keeps the effect engine liftable into
// go-blinkstick: no blinky package, and no other third-party code. The
// package's own _test.go files are allowed to import effecttest, its test
// support package, since that import does not travel with the package when
// it moves.
func TestImportsOnlyStdlibAndBlinkstick(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	more, _ := filepath.Glob("effecttest/*.go")
	files = append(files, more...)
	fset := token.NewFileSet()
	for _, f := range files {
		af, err := parser.ParseFile(fset, f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range af.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if path == "github.com/bawdo/go-blinkstick" {
				continue
			}
			if strings.HasSuffix(f, "_test.go") && path == "github.com/bawdo/blinky/internal/effect/effecttest" {
				continue
			}
			first, _, _ := strings.Cut(path, "/")
			if !strings.Contains(first, ".") {
				continue
			}
			t.Errorf("%s imports %s: internal/effect may import only the standard library and go-blinkstick", f, path)
		}
	}
}

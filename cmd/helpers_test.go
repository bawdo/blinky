package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/bawdo/blinky/internal/app"
	"github.com/bawdo/blinky/internal/effect/effecttest"
	"github.com/bawdo/blinky/internal/stick"
)

type result struct {
	out, err string
	code     int
}

// opts returns app options on ctl with a manual clock.
func opts(ctl stick.Controller) app.Options {
	return app.Options{Controller: ctl, Clock: effecttest.NewClock()}
}

// run executes blinky with args and returns its output and exit code.
func run(t *testing.T, o app.Options, args ...string) result {
	t.Helper()
	var out, errb bytes.Buffer
	o.Out, o.Err = &out, &errb
	root := newRootCmd(o)
	root.SetOut(&out)
	root.SetErr(&errb)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return result{out: out.String(), err: errb.String(), code: classifyError(err)}
}

// complete returns the choices shell completion offers for args.
func complete(t *testing.T, o app.Options, args ...string) []string {
	t.Helper()
	res := run(t, o, append([]string{"__complete"}, args...)...)
	var got []string
	for _, line := range strings.Split(res.out, "\n") {
		if strings.HasPrefix(line, ":") {
			break
		}
		if line != "" {
			name, _, _ := strings.Cut(line, "\t")
			got = append(got, name)
		}
	}
	return got
}

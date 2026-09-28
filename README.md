# blinky

A command line tool to control and manage BlinkStick Nano and Square LEDs, one stick or many at
once. Built on [go-blinkstick](https://github.com/bawdo/go-blinkstick). Unofficial, and not
affiliated with Agile Innovative.

## Requirements

- macOS. Linux will follow go-blinkstick.
- Go 1.26 or later, with cgo. Install the Xcode command line tools with `xcode-select --install`.

## Install

```
make install
```

or `go install github.com/bawdo/blinky@latest`.

## Quick start

```
blinky list                                            # what is plugged in, and the ID to use for each
blinky colour red                                      # the only stick goes red
blinky colour -a red blue                              # every stick: first half red, second half blue
blinky pulse -d desk vivid                             # breathe a random bright colour on the stick named desk
blinky blink green --on-time 1s --off-time 2s          # lit 1s, dark 2s, three times
blinky blink green --second-colour yellow --repeats 0  # green, yellow, until Ctrl-C
blinky morph green --from-colour yellow                # fade yellow to green
blinky morph green --from-colour yellow --loop         # back and forth until Ctrl-C
blinky police -a --duration 30s                        # red and blue for 30 seconds
blinky disco -a                                        # until Ctrl-C
blinky off -a
```

## Choosing sticks

`blinky list` shows an ID for every stick: its name if one is set, otherwise its serial.

| Flag | Meaning |
|---|---|
| `-d, --device <id>` | a stick by serial or name, repeatable. Serials match first |
| `-a, --all` | every attached stick blinky can drive |
| neither | the only attached stick, or an error listing the IDs if there are several |

`name` and `info-block` write to one stick at a time. Everything else works on a group.

## Colours

Hex (`ff8800`, `f80`), `r,g,b` (`255,136,0`), any of the 148 CSS names (`cornflowerblue`), or
`off`, `random` and `vivid`. In bash, quote hex with a leading `#` (`'#ff8800'`), or bash
treats it as a comment.

Given N colours, each stick splits its LEDs into N even groups. `red blue` gives a Nano one red
and one blue LED, and a Square four of each. With more colours than LEDs, a stick samples them
evenly.

## Commands

| Command | Does |
|---|---|
| `list` | list attached sticks |
| `info` | serial, model, firmware, manufacturer, product, name |
| `colour [<colour>...]` | read LEDs, or set them. `--led <i>` for one LED |
| `off` | turn LEDs off |
| `blink <colour>` | `--period 1s --repeats 3`, `--on-time`, `--off-time`, `--second-colour <colour>` |
| `pulse <colour>` | `--period 2s --repeats 3` |
| `morph <colour>` | `--fade 1s`, `--from-colour <colour>`, `--loop`, `--repeats 0` |
| `disco [<colour>...]` | `--min-period 200ms --max-period 2s --max-gap 1s` |
| `police` | `--period 1s`, `--alternate` |
| `name [<name>]` | read or set a stick's name. `--clear` removes it |
| `info-block <1\|2> [<data>]` | read or write an info block. `--hex`, `--clear` |
| `completion <shell>` | shell completion script |
| `version` | print the blinky version, commit and tag |

Commands that write LEDs take `--brightness <0-100>` and `--inverse`. Animated commands take
`--duration`, how long to run before stopping (0 means no limit). Ctrl-C or `--duration` running
out turns the LEDs off.

`morph` used to take `--duration` as its fade time. That is now `--fade`, and `--duration` means
stop after, as everywhere else.

`blink --on-time` and `--off-time` each default to half of `--period`. With `--second-colour`,
one repeat is colour, off, second colour, off. Times are drawn in 20ms frames, so on and off
times must be at least 20ms (off may be 0). `morph --loop` fades there and back once per
repeat and turns the LEDs off when it finishes. `--from-color` and `--second-color` work too.
Reads take `--json`. `name` and `info-block` write EEPROM, which wears out with heavy use.

## Shell completion

Completion knows each command's flags, the IDs of attached sticks and every colour name.

```
# zsh
blinky completion zsh > "${fpath[1]}/_blinky"
# bash
source <(blinky completion bash)
# fish
blinky completion fish | source
```

## Project layout

```
cmd/                  Cobra wiring. Flag parsing + delegation. Keep thin.
internal/
  app/                Orchestrator. Business logic. Unit-testable without Cobra.
  colour/             Colour arguments and the LED mapping rule.
  effect/             Frame engine, disco and police. Stdlib and go-blinkstick only.
  exitcode/           Sentinel errors mapped to documented exit codes.
  render/             Tables, fields and JSON output.
  settings/           Brightness and inverse, resolved from layers.
  stick/              Controller and Stick interfaces, hardware and fake.
  target/             --device and --all resolution.
  version/            Build identity (ldflags + runtime.debug fallback).
test/
  integration/        //go:build integration tests. Run via `make integration`.
```

A `cmd/foo.go` RunE should do flag checks and call one `app` method. Real work lives in
`internal/`.

## Exit codes

The CLI promises a stable exit-code contract. Sentinels live in
`internal/exitcode`; wrap with `fmt.Errorf("%w: ...", exitcode.ErrXxx, detail)`
and `cmd/root.go` translates to the right code.

| Code | Sentinel                   | Meaning                                                  |
|------|----------------------------|----------------------------------------------------------|
| 0    | (nil error)                | success, including a stop by Ctrl-C or `--duration`      |
| 1    | (unwrapped error)          | generic failure                                          |
| 2    | `exitcode.ErrInvalidArgs`  | invalid arguments / bad user input                       |
| 3    | `exitcode.ErrPrerequisite` | prerequisite missing (env/dep/state)                     |
| 4    | `exitcode.ErrNotFound`     | no sticks attached, `--device` matched nothing, or a name is on two sticks |
| 5    | `exitcode.ErrBusy`         | another program holds a stick you asked for              |
| 6    | `exitcode.ErrPartial`      | a group command worked on some sticks and failed on others |

Add codes by extending `internal/exitcode/exitcode.go` and updating both
this table and the test in `internal/exitcode/exitcode_test.go`.

## Integration tests

Hardware tests live in `test/integration/` behind `//go:build integration`. They use whatever
sticks are plugged in and skip when there are none, so CI (which has no sticks) skips them.

```
make integration
```

They write LEDs only. **Tests must never write EEPROM** (names or info blocks) on a real stick;
cover those against the fake in `internal/stick/sticktest`.

## Development

Run the local pre-CI gauntlet before pushing:

```
make pre-ci
```

Auto-fix formatting along the way:

```
make pre-ci-fix
```

Install a local build into `$GOBIN`:

```
make install
```

## TODO

- **Config file.** `~/.config/blinky/config.toml` with default brightness and inverse, plus
  per-stick overrides keyed by serial. It becomes a layer beneath the flags in
  `internal/settings`, so command line flags always take precedence.
- **Daemon mode.** A background process that owns the sticks, reached through a client that
  implements `stick.Controller`, so settings persist and effects can change while running.
- **Linux**, once go-blinkstick supports it.

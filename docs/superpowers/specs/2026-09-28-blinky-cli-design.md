# blinky CLI design

Date: 2026-09-28
Status: design awaiting review. go-blinkstick v0.3.0 is released

## Purpose

`blinky` is a command line tool to control and manage one or more BlinkSticks. It exposes every
feature of [go-blinkstick](https://github.com/bawdo/go-blinkstick), adds group control across
sticks with different LED counts, and adds two long-running effects: disco and police.

## Scope

- Devices: BlinkStick Nano (2 LEDs) and Square (8 LEDs).
- Platform: macOS only, as go-blinkstick is. Linux comes later with the library.
- Process model: foreground only. Every command runs, does its work and exits. Long-running
  effects run until Ctrl-C or `--duration`.
- Settings: command line flags only. No config file in v1.

### Out of scope for v1, designed for

- **Daemon mode.** All hardware access goes through the `stick.Controller` interface, so a daemon
  client can implement it later without changing commands.
- **Config file.** Settings are resolved from ordered layers. v1 has one layer (flags). A config
  file layer goes beneath it later, and flags always take precedence. Recorded as a TODO in
  `README.md`.
- **Moving the effect engine into go-blinkstick.** `internal/effect` imports only the standard
  library and go-blinkstick, enforced by a test.

## Prerequisites

- go-blinkstick **v0.3.0**, which adds `ColourNames() []string` (148 CSS names, sorted) and
  `RGB.Hex() string` (`#rrggbb`). Released 2026-09-28.
- Go 1.26 or later (go-blinkstick requires it). `go.mod` and CI move from 1.25.

## CLI grammar

Rules that every command follows:

1. `blinky <command> [args] [flags]`, one level deep. Only Cobra's `completion` command nests.
2. Sticks are chosen with `-d`/`-a`, never positionally.
3. **Stored things are nouns**: no argument reads, an argument writes, `--clear` clears
   (`colour`, `name`, `info-block`).
4. **Actions are verbs**: `list`, `info`, `off`, `blink`, `pulse`, `morph`, `disco`, `police`.
5. **Reads work on a group. Writes use the command's write scope.**
6. `--duration` always means how long the command runs (0 means until stopped). `--period` always
   means one cycle.
7. Every read command accepts `--json`.
8. Colours are positional arguments wherever a command takes colours.

## Targeting

| Flag | Meaning |
|---|---|
| `-d, --device <id>` | A stick by serial or name. Repeatable. Matches serial exactly first, then name exactly |
| `-a, --all` | Every attached stick |
| neither | The only attached stick. With two or more, error listing the IDs to choose from |

A stick's **ID** is its name if one is set, otherwise its serial. `list` shows it first.

Every command declares a **scope**:

| Scope | Meaning |
|---|---|
| `one` | Exactly one stick. `--all` or several `-d` is an error (exit 2) |
| `group` | One or more sticks |

Scope is declared once per command in `cmd/`, enforced by `internal/target`, printed in `--help`
("Targets: one stick" or "Targets: one or more sticks") and used by completion (a `one` command
never offers `--all`). For nouns, reads are always `group` and writes use the scope in the command
table.

## Global flags

| Flag | Default | Meaning |
|---|---|---|
| `--brightness <0-100>` | `100` | Percent. Mapped to `SetBrightnessLimit(round(p * 255 / 100))` |
| `--inverse` | off | `SetInverse(true)` on every target |

## Commands

| Command | Write scope | Library features | Flags (default) |
|---|---|---|---|
| `list` | n/a | `ListNamed` | `--json` |
| `info` | n/a (reads group) | `Info`, `Name` | `--json` |
| `colour [<colour>...]` | group | `Frame`, `LED`, `SetFrame`, `SetAll`, `SetLED` | `--led <i>`, `--json` |
| `name [<name>]` | one | `Name`, `SetName` | `--clear`, `--json` |
| `info-block <1\|2> [<data>]` | one | `InfoBlock`, `SetInfoBlock` | `--hex`, `--clear`, `--json` |
| `off` | group | `Off` | |
| `blink <colour>` | group | `Blink` | `--period` (1s), `--repeats` (3), `--duration` (0) |
| `pulse <colour>` | group | `Pulse` | `--period` (2s), `--repeats` (3), `--duration` (0) |
| `morph <colour>` | group | `Morph` | `--duration` (1s, the fade time) |
| `disco` | group | effect engine, `RandomVivid` | `--min-period` (200ms), `--max-period` (2s), `--max-gap` (1s), `--palette` (none), `--duration` (0) |
| `police` | group | effect engine | `--period` (1s), `--alternate` (off), `--duration` (0) |
| `version` | n/a | | exists |
| `completion <shell>` | n/a | | Cobra built-in: bash, zsh, fish |

`color` is a hidden alias of `colour`.

### Command details

- **`colour`** with no colours prints each stick's LEDs. With colours it applies the mapping rule.
  `--led <i>` reads or writes one physical LED. With several sticks it errors (exit 2) unless the
  index exists on every one. `--led` takes exactly one colour.
- **`name <name>`** refuses: a name another attached stick already has, a name shaped like a
  serial (`^BS\d+-\d+\.\d+$`), and names with control characters. The library rules also apply
  (UTF-8, no NUL, at most 32 bytes).
- **`info-block <n> <data>`** writes `data` as text, or as bytes with `--hex` (`68656c6c6f` or
  `68 65 6c 6c 6f`). Writing block 1 prints a warning that it holds the stick's name.
- **`blink` and `pulse`**: `--repeats 0` means forever. They stop at whichever of `--repeats` or
  `--duration` comes first. `blink` and `pulse` repeat by calling the library method with
  `repeats` 1 in a loop, so `--duration` can cut in between.
- **EEPROM.** `name` and `info-block` writes carry an EEPROM wear warning in `--help`.

## Colours

A colour argument accepts everything `ParseRGB` does:

- hex `#rgb` or `#rrggbb`, with or without `#`
- `r,g,b` decimal
- the 148 CSS colour names, any case

plus three keywords handled in `internal/colour`:

| Keyword | Value |
|---|---|
| `off` | `RGB{}` |
| `random` | `RandomRGB()`, picked fresh for every LED group on every stick |
| `vivid` | `RandomVivid()`, picked fresh for every LED group on every stick |

Docs and examples use bare hex (`ff0000`) or quoted hex (`'#ff0000'`), because bash treats an
unquoted leading `#` as a comment.

### Mapping rule

Given N colours and a stick with L LEDs, LED `i` gets colour `floor(i * N / L)`. The same rule
applies to one stick or many. When N > L the stick samples evenly spaced colours.

| Colours | Nano (2) | Square (8) |
|---|---|---|
| `red` | red, red | all red |
| `red blue` | red, blue | 0-3 red, 4-7 blue |
| `red green blue white` | red, blue | 0-1 red, 2-3 green, 4-5 blue, 6-7 white |
| 8 colours | colours 0 and 4 | one each |

## Effects

Both effects run until stopped or until `--duration` ends.

### Disco

Every LED on every stick runs an independent loop:

1. Pick a colour: `RandomVivid()`, or a random entry from `--palette` if given
   (comma-separated, any colour format).
2. Pulse it (fade up, fade down) over a random period in `[--min-period, --max-period]`.
3. Stay dark for a random gap in `[0, --max-gap]`.
4. Repeat.

A hidden `--seed` flag makes the output repeatable.

### Police

The period is split into two halves. The red half and blue half of each stick follow the mapping
rule with two groups (Nano LED 0 and 1; Square 0-3 and 4-7).

| Phase | Red half | Blue half |
|---|---|---|
| First half of period | off to red | blue to off |
| Second half | red to off | off to blue |

All sticks share one clock. With `--alternate`, every second stick in serial order swaps halves,
so neighbours run in opposite phase.

## Architecture

```
cmd/                 Cobra wiring: scope, flags, completion funcs. RunE ~3 lines.
internal/app/        Orchestration: resolve targets, apply settings, run against each stick,
                     aggregate per-stick errors.
internal/stick/      Controller and Stick interfaces. v1 implementation wraps go-blinkstick.
                     In-memory fake for tests. Future daemon client implements the same.
internal/target/     Pure resolution of -d/-a against the stick list, and scope checks.
internal/settings/   Settings{Brightness, Inverse} resolved from ordered layers (v1: flags).
internal/colour/     Colour argument parsing, keywords and the mapping rule.
internal/effect/     Frame engine: Effect interface, Runner, Disco, Police.
                     Imports only stdlib and go-blinkstick.
internal/render/     Tables, info blocks and JSON output.
internal/exitcode/   Sentinels (exists, extended).
internal/version/    Exists.
```

### Effect engine

- An `Effect` produces the frame for a stick given elapsed time and its LED count. Disco keeps
  per-LED state and draws from an injected RNG.
- The `Runner` ticks every 20ms (matching the library's effect step). On each tick it asks the
  effect for each stick's frame and writes it with one `SetFrame` call.
- The runner writes to a `Sink` interface (`LEDs() int`, `SetFrame([]blinkstick.RGB) error`), not
  to blinky types. `*blinkstick.Device` satisfies it through a one-line adapter for `LEDs`.
- Clock and RNG are injected so tests are deterministic.

### Data flow

`blinky colour -a red blue`:

1. `cmd` parses flags and calls `app`.
2. `app` calls `Controller.List`, then `target.Resolve` (checks scope).
3. `Controller.Open` opens the chosen sticks.
4. `settings` applies brightness and inverse to each.
5. `colour.Map` builds each stick's frame from its LED count, and `SetFrame` writes it.
6. Every stick is closed.

For `disco` and `police`, step 5 hands the sticks to `effect.Runner` with the command context.
Ctrl-C and `--duration` both cancel it.

## Errors and exit codes

| Code | Sentinel | When |
|---|---|---|
| 0 | nil | Success, including a stop by Ctrl-C or `--duration` |
| 1 | unwrapped | Other failure |
| 2 | `ErrInvalidArgs` | Bad colour, duration or index. Scope violation. Invalid, duplicate or serial-shaped name |
| 3 | `ErrPrerequisite` | Prerequisite missing (existing) |
| 4 | `ErrNotFound` (new) | No sticks attached for a command that needs one, `-d` matched nothing, a name is on two sticks |
| 5 | `ErrBusy` (new) | Another process holds a requested stick |
| 6 | `ErrPartial` (new) | A group command succeeded on some sticks and failed on others |

- **Best effort in groups.** A group command acts on every target that opens and works, reports
  each failure on stderr, and exits 6 if at least one stick succeeded. If none succeeded it exits
  with the code of the underlying failure.
- **Busy name resolution.** If `-d <name>` finds nothing and some sticks are busy, the error says
  so and lists the busy serials: `desk not found, 1 stick busy: BS072777-3.0, try its serial`.
- **Output.** Results to stdout. Errors to stderr, one line per stick, prefixed with its ID.
- **Stopping.** When an animated command (`blink`, `pulse`, `morph`, `disco`, `police`) is cut
  short by Ctrl-C, or by `--duration` (for `blink` and `pulse`, before `--repeats` completes),
  blinky turns its LEDs off using a fresh short timeout and exits 0. `disco` and `police` always
  end this way. A command that finishes on its own leaves the LEDs as the library does: `blink`
  and `pulse` end off, `morph` ends on its target colour (its `--duration` is the fade, not a
  cut-off). A second Ctrl-C exits immediately.
- **Unplugged during an effect.** The runner keeps animating the other sticks. It prints one
  warning when a stick drops out and one when it returns, and retries the missing stick at most
  once a second. A disconnect never ends the run.

## Output

### `list`

```
ID            NAME  SERIAL        MODEL    LEDS  STATUS
desk          desk  BS072777-3.0  Nano     2     ok
BS073788-3.1  -     BS073788-3.1  Square   8     ok
BS000001-3.0  ?     BS000001-3.0  Square   8     busy
BS000002-2.0  -     BS000002-2.0  unknown  -     unsupported
```

- Rows are sorted by serial. This is also the order `--alternate` uses.
- `-` means no name. `?` means the name could not be read (busy), and the ID falls back to the
  serial.
- With no sticks: `No BlinkSticks attached.` on stderr, exit 0. `--json` prints `[]`.

### Table formatting

- Column widths are computed from the content on every run.
- Width is terminal display width (`github.com/mattn/go-runewidth`), not bytes or runes, so
  CJK and emoji names stay aligned.
- Control characters are shown escaped (`desk\tone`) so a name cannot break the table or send
  escape sequences to the terminal.
- One renderer in `internal/render` is used for every table and for the `info` label column.

### `info`

```
desk
  Serial:        BS072777-3.0
  Model:         Nano (2 LEDs)
  Firmware:      3.0
  Manufacturer:  Agile Innovative Ltd
  Product:       BlinkStick Nano
  Name:          desk
```

### Noun reads

Always one line per stick, prefixed with its ID, even for a single stick:

```
desk: #ff0000 #0000ff
desk: desk
desk: hello  (68 65 6c 6c 6f)
```

### JSON

Every read supports `--json`. The output is always an array, one element per stick, using these
field names: `id`, `name`, `serial`, `model`, `leds`, `firmware`, `manufacturer`, `product`,
`status`, `colours` (hex strings), `info_block` (`{"text": ..., "hex": ...}`).

## Shell completion

Generated by `blinky completion bash|zsh|fish`. Each command registers its own completion.

| Position | Completes |
|---|---|
| command | subcommands |
| flags | only that command's flags (`police --<tab>` never shows disco flags) |
| `-d <tab>` | live stick IDs from the attached hardware |
| colour arguments, `--palette` (after each comma too) | the 148 CSS names and the keywords. zsh and fish show the hex as a description |
| duration flags | `500ms`, `1s`, `5s`, `30s` |
| `info-block <tab>` | `1`, `2` |

`<tab><tab>` on a colour argument lists every name.

## Testing

- **Unit tests, no hardware.** They run against an in-memory fake `stick.Controller` that records
  every frame. Table tests cover target resolution, scope, the mapping rule and keywords, settings
  layering, exit codes and per-stick error aggregation.
- **Effects.** A fake clock and a seeded RNG let tests assert exact frames at given times,
  including police `--alternate` opposite phase across two sticks.
- **Commands.** They run through Cobra against the fake. Completion is tested through Cobra's
  hidden `__complete` command.
- **Isolation.** A test parses `internal/effect` imports and fails on anything other than stdlib
  or go-blinkstick.
- **Rendering.** Tables are tested with ASCII, CJK, emoji and control character names.
- **Hardware** (`test/integration`, `make integration`). These skip unless a stick is attached and
  use whatever sticks are present. They write LEDs only. **They never write EEPROM**: `name` and
  `info-block` writes are tested against the fake only.

## CI changes

- Go `1.25.x` becomes a `1.26` and `1.27` matrix, matching go-blinkstick.
- The integration job moves from `ubuntu-latest` to `macos-latest`. go-blinkstick does not
  support Linux, and there it would skip for lack of hardware anyway.

## README changes

- Usage, targeting, colours, commands, exit codes (4, 5, 6 added), completion setup.
- TODO section at the bottom: a config file (`~/.config/blinky/config.toml`) with global
  defaults and per-stick overrides keyed by serial, beneath flags in precedence. Daemon mode.

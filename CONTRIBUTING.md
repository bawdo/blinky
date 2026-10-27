# Contributing

Thanks for taking an interest. blinky is small and maintained by one
person, so the process is light.

## Before you start

Open an issue for anything bigger than a typo or a small fix, so the
approach can be agreed before you spend time on it. Security problems go
through [SECURITY.md](SECURITY.md), not the issue tracker.

## Making a change

1. Fork and branch from `main`.
2. Run `make pre-ci` before pushing. It mirrors the GitHub Actions checks.
   `make pre-ci-fix` fixes formatting along the way.
3. Update the [README](README.md) if the change alters behaviour a user
   would notice, such as a command, flag or exit code.
4. Open a pull request. CI must pass before it can be merged. There is no
   reviewer requirement, but expect questions.

`scripts/history-scan.sh` is run before anything is pushed to a public remote.

Commits are signed and written in the past tense, with a subject of 60
characters or fewer. Use Australian English in prose (licence, behaviour,
organise).

## Testing against hardware

Unit tests run against the fake controller in `internal/stick/sticktest`
and need no hardware. `make integration` runs the tests in
`test/integration/` against whatever sticks are plugged in, and skips when
there are none, so CI skips them.

Integration tests write LEDs only. **Tests must never write EEPROM** (names
or info blocks) on a real stick, because EEPROM wears out and a stray write
changes someone's stick. Cover `name` and `info-block` against the fake
instead.

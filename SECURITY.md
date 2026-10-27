# Security policy

blinky is a command line tool that drives BlinkStick Nano and Square USB
LEDs. It talks to the sticks as USB HID devices through
[go-blinkstick](https://github.com/bawdo/go-blinkstick) (cgo and hidapi).
Most commands only set LEDs. Two, `name` and `info-block`, write a stick's
EEPROM, and they act on exactly one stick at a time, chosen with
`--device` or because it is the only one attached. blinky has no network
code, never runs `sudo` or starts another program, and runs with the
privileges of the user who invoked it. It reads no config file and writes
nothing to the host's filesystem: settings come from command line flags
only, and `blinky completion` prints its script to standard output for you
to redirect where you choose.

Bugs that weaken that boundary are security bugs: writing EEPROM from a
command documented not to (anything other than `name` and `info-block`, or
a read with no value given); acting on a stick the user did not select, or
on more than one stick when a command promises one; data read from a stick
(its name, serial or info blocks) being able to inject control sequences
into the terminal or the shell completion output; blinky starting another
program or passing untrusted input to one; or any path where blinky gains
or asks for elevated privileges. Behaviour the README already documents (for
example, that EEPROM wears out with heavy use) is not a vulnerability,
though a better fix is always welcome.

Bugs in go-blinkstick itself are welcome here too, or at
<https://github.com/bawdo/go-blinkstick/security/advisories/new>.

## Supported versions

There are no tagged releases yet. blinky is installed from source with
`go install` or `make install`, and only the head of `main` receives fixes.
There are no maintenance branches.

## Reporting a vulnerability

Please do not open a public issue for a security problem.

Use GitHub's private reporting form:
<https://github.com/bawdo/blinky/security/advisories/new>. It goes
straight to the maintainer and nobody else can see it.

Include what you can of: the output of `blinky version` and `blinky list`,
your OS and its version, the stick model (Nano, Square or other) and its
firmware, the steps involved, what you expected, and what happened. A proof
of concept helps but is not required.

You will get an acknowledgement within seven days. A fix, or an explanation
of why the report is not being treated as a vulnerability, follows as a
published advisory and a fix on `main`. Credit is given in the advisory
unless you ask otherwise.

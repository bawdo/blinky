#!/usr/bin/env bash
#
# Tests for scripts/outdated.sh. Each case builds a throwaway project
# directory (go.mod, CI workflow, Makefile) and puts a stub `go` first on
# PATH that answers the script's `go list` calls from fixture files, so no
# case touches the network. Output is captured, so the script runs without
# a terminal: it reports and never offers upgrades. Prints "ok <name>" per
# case and exits non-zero on the first failure.

set -euo pipefail

SCRIPTS="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(mktemp -d "${TMPDIR:-/tmp}/outdated_test.XXXXXX")"
trap 'rm -rf "$ROOT"' EXIT

OLD="2020-01-01T00:00:00Z"
FRESH="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# The stub go. STUB_DIR points at the fixture's answers:
#   all.json          go list -m -u -json all (all.fail makes it fail)
#   latest/<mod>.json go list -m -json <mod>@latest, "/" in <mod> as "_"
#   go-latest         go list -m -f '{{.Version}}' go@latest
mkdir "$ROOT/bin"
cat > "$ROOT/bin/go" <<'STUB'
#!/usr/bin/env bash
set -euo pipefail
case "$*" in
    "list -m -u -json all")
        if [ -e "$STUB_DIR/all.fail" ]; then
            echo "go: stub lookup failed" >&2
            exit 1
        fi
        cat "$STUB_DIR/all.json"
        ;;
    "list -m -json "*@latest)
        mod="${4%@latest}"
        f="$STUB_DIR/latest/${mod//\//_}.json"
        [ -f "$f" ] || exit 1
        cat "$f"
        ;;
    "list -m -f {{.Version}} go@latest")
        cat "$STUB_DIR/go-latest"
        ;;
    *)
        echo "stub go: unexpected call: $*" >&2
        exit 2
        ;;
esac
STUB
chmod +x "$ROOT/bin/go"

OUT=""
ERR=""
CODE=0

fail() {
    echo "FAIL $1: $2" >&2
    echo "  exit: $CODE" >&2
    echo "  stdout: $OUT" >&2
    echo "  stderr: $ERR" >&2
    exit 1
}

# has NAME TEXT / lacks NAME TEXT: asserts on stdout as a fixed string.
has() { grep -q -F -- "$2" <<< "$OUT" || fail "$1" "stdout lacks: $2"; }
lacks() { ! grep -q -F -- "$2" <<< "$OUT" || fail "$1" "stdout has: $2"; }

# latest NAME MODULE VERSION TIME: the stub's answer for MODULE@latest.
latest() {
    printf '{"Path":"%s","Version":"%s","Time":"%s"}\n' "$2" "$3" "$4" \
        > "$ROOT/$1/stub/latest/${2//\//_}.json"
}

# fixture NAME: a project where every module and pin is up to date.
fixture() {
    local dir="$ROOT/$1"
    mkdir -p "$dir/.github/workflows" "$dir/stub/latest"
    printf 'module example.com/app\n\ngo 1.26\n' > "$dir/go.mod"
    cat > "$dir/.github/workflows/ci.yml" <<'YAML'
env:
  VULN_VERSION: v1.9.0
jobs:
  lint:
    steps:
      - uses: actions/setup-go@abc # v7
        with:
          go-version: '1.26.x'
      - uses: golangci/golangci-lint-action@abc # v9
        with:
          version: v2.15.0
  test:
    strategy:
      matrix:
        go: ['1.26.x', '1.27.x']
YAML
    printf 'VULN_VERSION := v1.9.0\n' > "$dir/Makefile"
    cat > "$dir/stub/all.json" <<'JSON'
{"Path":"example.com/app","Main":true}
{"Path":"example.com/current","Version":"v1.0.0"}
JSON
    latest "$1" github.com/golangci/golangci-lint/v2 v2.15.0 "$OLD"
    latest "$1" golang.org/x/vuln v1.9.0 "$OLD"
    echo "1.26.3" > "$dir/stub/go-latest"
}

# run NAME: runs the script inside the fixture, with no terminal.
run() {
    local dir="$ROOT/$1" errf="$ROOT/$1.err"
    CODE=0
    OUT="$(cd "$dir" && PATH="$ROOT/bin:$PATH" STUB_DIR="$dir/stub" \
        bash "$SCRIPTS/outdated.sh" < /dev/null 2> "$errf")" || CODE=$?
    ERR="$(cat "$errf")"
}

test_rejects_bad_cooldown() {
    fixture cooldown
    CODE=0
    OUT="$(cd "$ROOT/cooldown" && COOLDOWN_DAYS=abc bash "$SCRIPTS/outdated.sh" 2>&1)" || CODE=$?
    [[ $CODE -eq 1 ]] || fail rejects_bad_cooldown "expected exit 1"
    has rejects_bad_cooldown "COOLDOWN_DAYS must be a whole number"
    echo "ok rejects_bad_cooldown"
}

test_reports_all_up_to_date() {
    fixture current
    run current
    [[ $CODE -eq 0 ]] || fail reports_all_up_to_date "expected exit 0"
    has reports_all_up_to_date "All direct modules are up to date."
    has reports_all_up_to_date "golangci-lint  v2.15.0    up to date"
    has reports_all_up_to_date "govulncheck    v1.9.0     up to date"
    has reports_all_up_to_date "up to date (latest 1.26.3)"
    lacks reports_all_up_to_date "pins disagree"
    lacks reports_all_up_to_date "Not on a terminal"
    echo "ok reports_all_up_to_date"
}

test_lists_outdated_direct_modules() {
    fixture modules
    cat > "$ROOT/modules/stub/all.json" <<JSON
{"Path":"example.com/app","Main":true}
{"Path":"example.com/old","Version":"v1.0.0","Update":{"Version":"v1.2.0","Time":"2020-01-01T00:00:00.123Z"}}
{"Path":"example.com/new","Version":"v2.0.0","Update":{"Version":"v2.1.0","Time":"$FRESH"}}
{"Path":"example.com/dep","Version":"v0.1.0","Indirect":true,"Update":{"Version":"v0.2.0","Time":"$OLD"}}
JSON
    run modules
    [[ $CODE -eq 0 ]] || fail lists_outdated_direct_modules "expected exit 0"
    grep -q -E '^  example\.com/old +v1\.0\.0 +v1\.2\.0 +[0-9]+d$' <<< "$OUT" \
        || fail lists_outdated_direct_modules "old module row missing or mis-aged"
    grep -q -E '^  example\.com/new +v2\.0\.0 +v2\.1\.0 +0d, too fresh$' <<< "$OUT" \
        || fail lists_outdated_direct_modules "fresh module not flagged"
    lacks lists_outdated_direct_modules "example.com/dep"
    has lists_outdated_direct_modules "1 indirect module(s) also have updates"
    has lists_outdated_direct_modules "Not on a terminal, so not offering upgrades."
    echo "ok lists_outdated_direct_modules"
}

test_warns_deprecated_and_retracted() {
    fixture warn
    cat > "$ROOT/warn/stub/all.json" <<'JSON'
{"Path":"example.com/app","Main":true}
{"Path":"example.com/gone","Version":"v1.0.0","Deprecated":"use example.com/new"}
{"Path":"example.com/bad","Version":"v1.3.0","Retracted":["broken build","bad tag"]}
JSON
    run warn
    has warns_deprecated_and_retracted "! example.com/gone v1.0.0: deprecated: use example.com/new"
    has warns_deprecated_and_retracted "! example.com/bad v1.3.0: retracted: broken build; bad tag"
    echo "ok warns_deprecated_and_retracted"
}

test_reads_vuln_pin_from_env_and_makefile() {
    fixture vulnenv
    latest vulnenv golang.org/x/vuln v1.10.0 "$OLD"
    run vulnenv
    has reads_vuln_pin_from_env_and_makefile "govulncheck    v1.9.0     -> v1.10.0"
    has reads_vuln_pin_from_env_and_makefile "pinned in .github/workflows/ci.yml:2, Makefile:1"
    lacks reads_vuln_pin_from_env_and_makefile "pins disagree"
    echo "ok reads_vuln_pin_from_env_and_makefile"
}

test_flags_vuln_pin_drift() {
    fixture drift
    printf 'VULN_VERSION := v1.8.0  # older copy\n' > "$ROOT/drift/Makefile"
    run drift
    has flags_vuln_pin_drift "! pins disagree: v1.9.0 in .github/workflows/ci.yml, v1.8.0 in Makefile"
    echo "ok flags_vuln_pin_drift"
}

test_falls_back_to_literal_vuln_pin() {
    fixture literal
    sed -i.bak '/VULN_VERSION/d' "$ROOT/literal/.github/workflows/ci.yml"
    printf '      - run: go run golang.org/x/vuln/cmd/govulncheck@v1.8.0 ./...\n' \
        >> "$ROOT/literal/.github/workflows/ci.yml"
    rm "$ROOT/literal/Makefile"
    run literal
    has falls_back_to_literal_vuln_pin "govulncheck    v1.8.0     -> v1.9.0"
    has falls_back_to_literal_vuln_pin "pinned in .github/workflows/ci.yml:15"
    echo "ok falls_back_to_literal_vuln_pin"
}

test_reports_missing_vuln_pin() {
    fixture novuln
    sed -i.bak '/VULN_VERSION/d' "$ROOT/novuln/.github/workflows/ci.yml"
    run novuln
    has reports_missing_vuln_pin "govulncheck    pin not found in .github/workflows/ci.yml"
    echo "ok reports_missing_vuln_pin"
}

test_ignores_install_hint_in_pre_ci() {
    fixture hint
    mkdir "$ROOT/hint/scripts"
    echo 'echo "Install: go install golang.org/x/vuln/cmd/govulncheck@latest"' \
        > "$ROOT/hint/scripts/pre-ci-check.sh"
    latest hint golang.org/x/vuln v1.10.0 "$OLD"
    run hint
    has ignores_install_hint_in_pre_ci "pinned in .github/workflows/ci.yml:2, Makefile:1"
    lacks ignores_install_hint_in_pre_ci "scripts/pre-ci-check.sh"
    echo "ok ignores_install_hint_in_pre_ci"
}

test_go_line_notes_test_matrix() {
    fixture matrix
    echo "1.27.4" > "$ROOT/matrix/stub/go-latest"
    run matrix
    has go_line_notes_test_matrix "1.26.x     -> 1.27.4     newer release line"
    has go_line_notes_test_matrix "1.27.x is already in the test matrix (.github/workflows/ci.yml:15)"
    has go_line_notes_test_matrix "go.mod says go 1.26"
    echo "ok go_line_notes_test_matrix"
}

test_go_line_without_matrix_entry() {
    fixture nomatrix
    echo "1.27.4" > "$ROOT/nomatrix/stub/go-latest"
    sed -i.bak "s/go: \['1.26.x', '1.27.x'\]/go: ['1.26.x']/" "$ROOT/nomatrix/.github/workflows/ci.yml"
    run nomatrix
    has go_line_without_matrix_entry "newer release line"
    lacks go_line_without_matrix_entry "already in the test matrix"
    echo "ok go_line_without_matrix_entry"
}

test_go_list_failure_exits() {
    fixture listfail
    touch "$ROOT/listfail/stub/all.fail"
    run listfail
    [[ $CODE -eq 1 ]] || fail go_list_failure_exits "expected exit 1"
    grep -q -F "outdated: go list failed:" <<< "$ERR" || fail go_list_failure_exits "stderr lacks failure"
    grep -q -F "stub lookup failed" <<< "$ERR" || fail go_list_failure_exits "go's stderr not passed on"
    echo "ok go_list_failure_exits"
}

test_rejects_bad_cooldown
test_reports_all_up_to_date
test_lists_outdated_direct_modules
test_warns_deprecated_and_retracted
test_reads_vuln_pin_from_env_and_makefile
test_flags_vuln_pin_drift
test_falls_back_to_literal_vuln_pin
test_reports_missing_vuln_pin
test_ignores_install_hint_in_pre_ci
test_go_line_notes_test_matrix
test_go_line_without_matrix_entry
test_go_list_failure_exits

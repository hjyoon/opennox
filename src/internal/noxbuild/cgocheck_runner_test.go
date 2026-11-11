package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStrictCGOCheckRunner(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX test runner")
	}
	versionData, err := os.ReadFile("../../../toolchain/go-version.txt")
	if err != nil {
		t.Fatal(err)
	}
	requiredGo := strings.TrimSpace(string(versionData))
	const fakeGo = `#!/bin/sh
set -eu
if [ "$1" = env ]; then
    case "$2" in
        GOVERSION)
            if [ "${GOEXPERIMENT:-}" = cgocheck2 ]; then
                printf '%s\n' "${NOX_TEST_STRICT_VERSION:-$NOX_TEST_PINNED_VERSION}"
            else
                printf '%s\n' "$NOX_TEST_PINNED_VERSION"
            fi ;;
        GOROOT)
            [ "${GOEXPERIMENT:-}" = '' ]
            printf '%s\n' "$NOX_TEST_TOOLCHAIN" ;;
        GOEXPERIMENT) printf '%s\n' "${NOX_TEST_REPORTED_EXPERIMENT-${GOEXPERIMENT:-}}" ;;
        CGO_ENABLED) printf '%s\n' "$CGO_ENABLED" ;;
        *) exit 90 ;;
    esac
    exit 0
fi
# The shell entry must delegate execution to the prepared-test driver.
exit 91
`
	const fakePreparedDriver = `#!/bin/sh
set -eu
[ "${GOROOT+x}" != x ]
[ "$GOTOOLCHAIN" = local ]
[ "$GOEXPERIMENT" = cgocheck2 ]
[ "$CGO_ENABLED" = 1 ]
[ "$CGO_CFLAGS_ALLOW" = '-f.*' ]
[ "$PKG_CONFIG_PATH" = 'test-pkgconfig' ]
[ "$(basename "$1")" = run-prepared-go-tests.py ]
[ -f "$1" ]
shift
[ "$1" = --go ]
[ "$2" = "$NOX_TEST_TOOLCHAIN/bin/go" ]
shift 2
[ "$1" = --label ]
[ "$2" = strict-cgocheck2 ]
shift 2
[ "$1" = --isolate-main-names ]
[ "$2" = -- ]
shift 2
printf 'prepared-args:\n'
printf '%s\n' "$@"
exit "${NOX_TEST_EXIT_CODE:-0}"
`
	tests := []struct {
		name string
		args []string
		env  []string
		code int
		want string
	}{
		{name: "defaults", want: "prepared-args:\n./...\n"},
		{name: "arguments", args: []string{"-run", "TestPointer", "-count=2", "./legacy"}, want: "prepared-args:\n-run\nTestPointer\n-count=2\n./legacy\n"},
		{name: "test failure", env: []string{"NOX_TEST_EXIT_CODE=7"}, code: 7, want: "prepared-args:\n./...\n"},
		{name: "wrong toolchain", env: []string{"NOX_TEST_STRICT_VERSION=go1.26.6"}, code: 2, want: "configuration mismatch"},
		{name: "experiment not applied", env: []string{"NOX_TEST_REPORTED_EXPERIMENT="}, code: 2, want: "configuration mismatch"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "toolchain with spaces")
			if err := os.MkdirAll(filepath.Join(root, "bin"), 0755); err != nil {
				t.Fatal(err)
			}
			goPath := filepath.Join(root, "bin", "go")
			if err := os.WriteFile(goPath, []byte(fakeGo), 0755); err != nil {
				t.Fatal(err)
			}
			pythonPath := filepath.Join(root, "bin", "prepared driver")
			if err := os.WriteFile(pythonPath, []byte(fakePreparedDriver), 0755); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("sh", append([]string{"../../../scripts/test-cgocheck2.sh"}, tc.args...)...)
			for _, v := range os.Environ() {
				key, _, _ := strings.Cut(v, "=")
				switch key {
				case "GO", "GOROOT", "GOTOOLCHAIN", "GOEXPERIMENT", "CGO_ENABLED", "CGO_CFLAGS_ALLOW", "PKG_CONFIG_PATH", "OPENNOX_TEST_PYTHON":
					continue
				}
				if !strings.HasPrefix(key, "NOX_TEST_") {
					cmd.Env = append(cmd.Env, v)
				}
			}
			cmd.Env = append(cmd.Env, "GO="+goPath, "NOX_TEST_TOOLCHAIN="+root, "NOX_TEST_PINNED_VERSION="+requiredGo,
				"OPENNOX_TEST_PYTHON="+pythonPath,
				"GOROOT=invalid-outer-root", "GOTOOLCHAIN=invalid-outer-toolchain", "GOEXPERIMENT=invalid-outer-experiment",
				"CGO_ENABLED=0", "CGO_CFLAGS_ALLOW=", "PKG_CONFIG_PATH=test-pkgconfig")
			cmd.Env = append(cmd.Env, tc.env...)
			out, err := cmd.CombinedOutput()
			if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != tc.code {
				t.Fatalf("exit status = %v, want %d: %v\n%s", cmd.ProcessState, tc.code, err, out)
			}
			if !strings.Contains(string(out), tc.want) {
				t.Fatalf("missing %q in output:\n%s", tc.want, out)
			}
			if tc.code == 2 && strings.Contains(string(out), "prepared-args:\n") {
				t.Fatalf("tests ran despite configuration mismatch:\n%s", out)
			}
		})
	}
}

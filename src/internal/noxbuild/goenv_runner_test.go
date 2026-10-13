package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGoRunnerUsesPinnedGoenv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX Go runner")
	}
	versionData, err := os.ReadFile("../../../toolchain/go-version.txt")
	if err != nil {
		t.Fatal(err)
	}
	required := strings.TrimSpace(string(versionData))
	const fakeGo = `#!/bin/sh
set -eu
[ "${GOROOT+x}" != x ]
[ "$GOTOOLCHAIN" = "$NOX_TEST_REQUIRED_GO" ]
[ "${GOEXPERIMENT:-}" = '' ]
[ "$1" = env ]
case "$2" in
    GOVERSION)
        if [ "$NOX_TEST_GO_KIND" = goenv ]; then
            printf '%s\n' "${NOX_TEST_GOENV_REPORTED_VERSION:-$NOX_TEST_REQUIRED_GO}"
        else
            printf '%s\n' "$NOX_TEST_REQUIRED_GO"
        fi ;;
    GOROOT) printf '%s\n' "$NOX_TEST_GO_KIND" ;;
    *) exit 81 ;;
esac
`
	const fakeGoenv = `#!/bin/sh
set -eu
printf '%s\n' "$@" >> "$NOX_TEST_GOENV_LOG"
[ "$#" = 2 ]
[ "$1" = prefix ]
[ "$2" = "${NOX_TEST_REQUIRED_GO#go}" ]
[ "$NOX_TEST_GOENV_INSTALLED" = 1 ] || exit 1
printf '%s\n' "$NOX_TEST_GOENV_PREFIX"
`
	cases := []struct {
		name         string
		explicit     bool
		noGoenv      bool
		missing      bool
		nonexec      bool
		wrongVersion bool
		want         string
		code         int
	}{
		{name: "installed version with spaces", want: "goenv"},
		{name: "explicit GO wins", explicit: true, want: "explicit"},
		{name: "version not installed", missing: true, want: "fallback"},
		{name: "goenv unavailable", noGoenv: true, want: "fallback"},
		{name: "nonexecutable installed candidate", nonexec: true, want: "fallback"},
		{name: "installed version mismatch", wrongVersion: true, want: "Go toolchain mismatch", code: 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			bin := filepath.Join(root, "path-bin")
			prefix := filepath.Join(root, "goenv toolchain with spaces")
			if err := os.MkdirAll(bin, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(prefix, "bin"), 0755); err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"dirname", "tr"} {
				path, err := exec.LookPath(name)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(path, filepath.Join(bin, name)); err != nil {
					t.Fatal(err)
				}
			}
			goScript := func(kind string) []byte {
				return []byte(strings.Replace(fakeGo, "set -eu\n", "set -eu\nNOX_TEST_GO_KIND="+kind+"\n", 1))
			}
			if err := os.WriteFile(filepath.Join(bin, "go"), goScript("fallback"), 0755); err != nil {
				t.Fatal(err)
			}
			explicitGo := filepath.Join(bin, "explicit-go")
			if err := os.WriteFile(explicitGo, goScript("explicit"), 0755); err != nil {
				t.Fatal(err)
			}
			mode := os.FileMode(0755)
			if tc.nonexec {
				mode = 0644
			}
			if err := os.WriteFile(filepath.Join(prefix, "bin", "go"), goScript("goenv"), mode); err != nil {
				t.Fatal(err)
			}
			if !tc.noGoenv {
				if err := os.WriteFile(filepath.Join(bin, "goenv"), []byte(fakeGoenv), 0755); err != nil {
					t.Fatal(err)
				}
			}
			logPath := filepath.Join(root, "goenv-prefix.log")
			cmd := exec.Command("sh", "../../../scripts/go.sh", "env", "GOROOT")
			for _, v := range os.Environ() {
				key, _, _ := strings.Cut(v, "=")
				switch key {
				case "GO", "GOROOT", "GOTOOLCHAIN", "GOEXPERIMENT", "PATH":
					continue
				}
				if !strings.HasPrefix(key, "NOX_TEST_") {
					cmd.Env = append(cmd.Env, v)
				}
			}
			installed := "1"
			if tc.missing {
				installed = "0"
			}
			cmd.Env = append(cmd.Env, "PATH="+bin, "GOROOT=invalid-outer-root", "GOTOOLCHAIN=invalid-outer-toolchain",
				"GOEXPERIMENT=invalid-outer-experiment", "NOX_TEST_REQUIRED_GO="+required,
				"NOX_TEST_GOENV_PREFIX="+prefix, "NOX_TEST_GOENV_INSTALLED="+installed, "NOX_TEST_GOENV_LOG="+logPath)
			if tc.explicit {
				cmd.Env = append(cmd.Env, "GO="+explicitGo)
			}
			if tc.wrongVersion {
				cmd.Env = append(cmd.Env, "NOX_TEST_GOENV_REPORTED_VERSION=go0.0.0")
			}
			out, err := cmd.CombinedOutput()
			if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != tc.code {
				t.Fatalf("exit=%v, want %d: %v\n%s", cmd.ProcessState, tc.code, err, out)
			}
			if !strings.Contains(string(out), tc.want) {
				t.Fatalf("output=%q, want %q", out, tc.want)
			}
			log, readErr := os.ReadFile(logPath)
			if tc.explicit || tc.noGoenv {
				if readErr == nil || !os.IsNotExist(readErr) {
					t.Fatalf("goenv called without authority or unavailable: %q, %v", log, readErr)
				}
			} else {
				if readErr != nil {
					t.Fatal(readErr)
				}
				if want := "prefix\n" + strings.TrimPrefix(required, "go") + "\n"; string(log) != want {
					t.Fatalf("goenv arguments=%q, want %q", log, want)
				}
			}
		})
	}
}

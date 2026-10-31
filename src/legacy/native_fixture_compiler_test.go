package legacy

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
)

// Windows reports the native DWORD process status instead of a POSIX signal.
// Callers still compare every byte of the original service prefix, including
// transport EFB. Compare abort against a direct call through the same compiler
// and CRT: UCRT's fast-fail status also represents unrelated failures, so it
// must not become an unconditional additional accepted status.
func nativeFixtureFaultIs(t *testing.T, exit *exec.ExitError, signal syscall.Signal) bool {
	t.Helper()
	if exit == nil {
		return false
	}
	if runtime.GOOS == "windows" {
		code := uint32(exit.ExitCode()) // 386 sign-extends NTSTATUS through int.
		switch signal {
		case syscall.SIGSEGV:
			return code == 0xc0000005 // STATUS_ACCESS_VIOLATION
		case syscall.SIGABRT:
			return code == nativeFixtureAbortCode(t)
		default:
			return false
		}
	}
	status, ok := exit.Sys().(syscall.WaitStatus)
	return ok && status.Signaled() && status.Signal() == signal
}

var nativeFixtureAbortControl struct {
	once sync.Once
	code uint32
}

func nativeFixtureAbortCode(t *testing.T) uint32 {
	t.Helper()
	nativeFixtureAbortControl.once.Do(func() {
		compiler := strings.Fields(os.Getenv("CC"))
		if len(compiler) == 0 {
			compiler = []string{"cc"}
		}
		binary := filepath.Join(t.TempDir(), "abort-control.exe")
		args := append(compiler[1:], "-std=c11", "-O2", "-Wall", "-Wextra", "-Werror", "-x", "c", "-", "-o", binary)
		cmd := exec.Command(compiler[0], nativeFixtureCompileArgs(t, args)...)
		cmd.Stdin = strings.NewReader("#include <stdio.h>\n#include <stdlib.h>\nint main(void) { fputs(\"native-abort-control\", stdout); fflush(stdout); abort(); }\n")
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("compile same-CRT abort control: %v\n%s", err, output)
		}
		output, err := exec.Command(binary).CombinedOutput()
		exit, ok := err.(*exec.ExitError)
		if !ok || string(output) != "native-abort-control" {
			t.Fatalf("direct same-CRT abort control: %v %q", err, output)
		}
		nativeFixtureAbortControl.code = uint32(exit.ExitCode())
		t.Logf("direct same-compiler/CRT abort status: %#x", nativeFixtureAbortControl.code)
	})
	return nativeFixtureAbortControl.code
}

// Match cgo_common.go's aliasing/overflow contract when compiling extracted
// production bodies. The forced test-only header preserves binary standard
// streams on Windows; it contributes no code on other hosts.
func nativeFixtureCompileArgs(t *testing.T, args []string) []string {
	t.Helper()
	header, err := filepath.Abs("testdata/native_stdio.h")
	if err != nil {
		t.Fatal(err)
	}
	out := append([]string{"-fno-strict-aliasing", "-fno-strict-overflow", "-include", header, "-I", filepath.Dir(header)}, args...)
	if runtime.GOOS == "windows" {
		// Match cgo_common.go's native Winsock linkage for byte-order services.
		out = append(out, "-lws2_32")
	}
	return out
}

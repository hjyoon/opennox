package legacy

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
)

// Windows reports the native DWORD process status instead of a POSIX signal.
// These exact codes were observed in both MinGW architectures. Callers still
// compare every byte of the original service prefix, including transport EFB.
func nativeFixtureFaultIs(exit *exec.ExitError, signal syscall.Signal) bool {
	if exit == nil {
		return false
	}
	if runtime.GOOS == "windows" {
		code := uint32(exit.ExitCode()) // 386 sign-extends NTSTATUS through int.
		switch signal {
		case syscall.SIGSEGV:
			return code == 0xc0000005 // STATUS_ACCESS_VIOLATION
		case syscall.SIGABRT:
			return code == 3 // MinGW CRT abort, at the declared transport service.
		default:
			return false
		}
	}
	status, ok := exit.Sys().(syscall.WaitStatus)
	return ok && status.Signaled() && status.Signal() == signal
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

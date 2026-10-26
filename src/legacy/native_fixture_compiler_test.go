package legacy

import (
	"path/filepath"
	"runtime"
	"testing"
)

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

package legacy

import (
	"path/filepath"
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
	return append([]string{"-fno-strict-aliasing", "-fno-strict-overflow", "-include", header}, args...)
}

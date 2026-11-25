package legacy

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestLegacyNetworkLogPrintf413D30(t *testing.T) {
	previous := NetworkLogPrint
	t.Cleanup(func() { NetworkLogPrint = previous })
	var messages []string
	NetworkLogPrint = func(text string) { messages = append(messages, text) }
	cfixture.NetworkLogPrintf413D30()
	if len(messages) != 1 || messages[0] != "queue 0037 ready" {
		t.Fatalf("network log = %q", messages)
	}
}

func TestLegacyFilePrintfNativeBuffer(t *testing.T) {
	handles.InitHigh()
	t.Cleanup(handles.Release)
	path := filepath.Join(t.TempDir(), "print.txt")
	file, err := binfile.BinfileOpen(path, binfile.WriteOnly)
	if err != nil {
		t.Fatal(err)
	}
	handle := NewFileHandle(file.File)
	func() {
		defer nox_fs_close(handle)
		for _, test := range []struct{ width, written int }{{4, 4}, {3000, 2047}} {
			if n := cfixture.FilePrintf(unsafe.Pointer(handle), test.width); n != test.written {
				t.Fatalf("width %d wrote %d bytes, want %d", test.width, n, test.written)
			}
		}
	}()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "0007"+strings.Repeat("0", 2047) {
		t.Fatalf("unexpected file contents of %d bytes", len(content))
	}
}

func TestLegacyWideFixedPrintfPreservesNativeFormatting(t *testing.T) {
	values := []uint64{
		0, 1, 0x8000000000000000, 0x8000000000000001,
		math.Float64bits(0.125), math.Float64bits(-0.135),
		math.Float64bits(0.999999999999), math.Float64bits(math.MaxFloat64),
		0x7ff0000000000000, 0xfff0000000000000, 0xfff8000000000000, 0x7ff0000000000001,
	}
	for rounding := range 4 {
		for _, digits := range []int{-1, 0, 1, 2, 5, 31, 62} {
			for _, value := range values {
				if got := cfixture.WideFixedPrintfMatches(value, digits, rounding); got != 1 {
					t.Errorf("bits=%016x precision=%d rounding=%d C comparison=%d", value, digits, rounding, got)
				}
			}
		}
	}
}

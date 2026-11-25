package legacy

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/libs/console"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestLegacyConsolePrintf450C00UsesBoundedWideBuffer(t *testing.T) {
	printer := &deathMessagePrinter4952E0{}
	con := console.New(printer)
	previous := GetConsole
	t.Cleanup(func() { GetConsole = previous })
	GetConsole = func() *console.Console { return con }

	for i, tc := range []struct{ format, want string }{
		{"%S %04d %s", "xp 0037 ź"},
		{"%S %0507d%s", "xp " + strings.Repeat("0", 505) + "37ź"},
		{"%S %0508d%s", "xp " + strings.Repeat("0", 506) + "37"},
		{"%S %0512d%s", "xp " + strings.Repeat("0", 508)},
		{"done %%", "done %"},
	} {
		format, free := CWString(tc.format)
		result := cfixture.ConsolePrintf450C00(byte(console.ColorRed), (*uint16)(unsafe.Pointer(format)), 37)
		free()
		if result != 1 || len(printer.lines) != i+1 || printer.colors[i] != console.ColorRed || printer.lines[i] != tc.want {
			t.Fatalf("format %q result %d, output %q", tc.format, result, printer.lines)
		}
	}
}

func TestLegacyWideNumberFormatterRetainsTerminatorAndGuards(t *testing.T) {
	format, free := CWString("%07d")
	defer free()
	buffer := [...]wchar2_t{0x1111, 0xa5a5, 0xa5a5, 0xa5a5, 0xa5a5, 0xa5a5, 0x2222}
	formatWideNumber(buffer[1:6], format, 37)
	want := [...]wchar2_t{0x1111, '0', '0', '0', '0', 0, 0x2222}
	if buffer != want {
		t.Fatalf("bounded integer output = %x, want %x", buffer, want)
	}
}

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

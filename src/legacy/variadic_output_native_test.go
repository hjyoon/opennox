package legacy

import (
	"context"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/opennox/libs/console"
	"github.com/opennox/opennox/v1/internal/binfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc/handles"
	"github.com/opennox/opennox/v1/legacy/internal/cfixture"
)

func TestLegacyFormatAbortPreservesDiagnosticBytes(t *testing.T) {
	const env = "NOX_FORMAT_ABORT_TEST_CODE"
	if value := os.Getenv(env); value != "" {
		code, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			t.Fatal(err)
		}
		cfixture.AbortUnsupportedWideFormat(uint16(code))
		return
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range []uint16{0, '?', 0x1e9} {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		cmd := exec.CommandContext(ctx, binary, "-test.run=^TestLegacyFormatAbortPreservesDiagnosticBytes$")
		cmd.Env = append(os.Environ(), env+"="+strconv.Itoa(int(code)))
		output, err := cmd.Output()
		timedOut := ctx.Err() != nil
		cancel()
		want := "Unhandled format character: '" + string([]byte{byte(code)}) + "'"
		if _, exited := err.(*exec.ExitError); timedOut || !exited || string(output) != want {
			t.Fatalf("format character %#x: stdout %q, want %q and CRT abort, got %v", code, output, want, err)
		}
	}
}

func TestLegacyConsolePrintf450C00UsesBoundedWideBuffer(t *testing.T) {
	printer := &deathMessagePrinter4952E0{}
	con := console.New(printer)
	previous := GetConsole
	t.Cleanup(func() { GetConsole = previous })
	GetConsole = func() *console.Console { return con }

	narrow := internCStr("xp")
	wide := internWStr("ź")
	for i, tc := range []struct{ format, want string }{
		{"%S %04d %s", "xp 0037 ź"},
		{"%S %0507d%s", "xp " + strings.Repeat("0", 505) + "37ź"},
		{"%S %0508d%s", "xp " + strings.Repeat("0", 506) + "37"},
		{"%S %0512d%s", "xp " + strings.Repeat("0", 508)},
		{"done %%", "done %"},
	} {
		format, free := CWString(tc.format)
		result := cfixture.ConsolePrintf450C00(byte(console.ColorRed), (*uint16)(unsafe.Pointer(format)), 37)
		consolePrintRed(format, formatValue{text: unsafe.Pointer(narrow)}, formatValue{number: 37},
			formatValue{text: unsafe.Pointer(wide)})
		free()
		if result != 1 || len(printer.lines) != 2*(i+1) ||
			printer.colors[2*i] != console.ColorRed || printer.colors[2*i+1] != console.ColorRed ||
			printer.lines[2*i] != tc.want || printer.lines[2*i+1] != tc.want {
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

func TestLegacyBoundedWideCEntriesMatchVariadic(t *testing.T) {
	for _, input := range []string{"done %%", "%c", "%07d", "%+4i", "%.12x", "%08X", "%o", "%u"} {
		format, free := CWString(input)
		t.Cleanup(free)
		for _, value := range []int32{0, 37, math.MinInt32, math.MaxInt32} {
			if !cfixture.BoundedWideNumberMatches((*uint16)(unsafe.Pointer(format)), value) {
				t.Errorf("number format %q, value %d differs from C variadic output", input, value)
			}
		}
	}
	first := [...]uint16{'A', '가', 0}
	second := [...]uint16{0xd83d, 0xde83, 0}
	for _, input := range []string{"done %%", "%s", "%s|%s", "%S|%s"} {
		format, free := CWString(input)
		t.Cleanup(free)
		for _, arg := range []*uint16{&second[0], nil} {
			if !cfixture.BoundedWideStringsMatch((*uint16)(unsafe.Pointer(format)), &first[0], arg) {
				t.Errorf("string format %q with second argument %p differs from C variadic output", input, arg)
			}
		}
	}
}

func TestLegacyFixedWideMatchesMixedVariadic(t *testing.T) {
	format, release := CWString("%S\t%012d|%s|%u%%")
	defer release()
	narrow := [...]byte{'M', 0xff, 0}
	wide := [...]uint16{'가', 0xd83d, 0xde83, 0}
	for _, args := range []struct {
		narrow *byte
		wide   *uint16
	}{
		{&narrow[0], &wide[0]},
		{nil, nil},
	} {
		var expected, actual [64]uint16
		for i := range expected {
			expected[i], actual[i] = 0xa5a5, 0xa5a5
		}
		length := cfixture.WideMixedPrintf(&expected[1], (*uint16)(unsafe.Pointer(format)),
			args.narrow, math.MinInt32, args.wide, math.MaxUint32)
		used := formatFixedWide((*wchar2_t)(unsafe.Pointer(&actual[1])), format,
			formatValue{text: unsafe.Pointer(args.narrow)}, formatValue{number: math.MinInt32},
			formatValue{text: unsafe.Pointer(args.wide)}, formatValue{number: -1})
		if int(used) != length || actual != expected {
			t.Fatalf("fixed output (%d, %x) differs from variadic output (%d, %x)", used, actual, length, expected)
		}
	}
}

func TestLegacyNetworkLogPrintf413D30(t *testing.T) {
	previous := NetworkLogPrint
	t.Cleanup(func() { NetworkLogPrint = previous })
	var messages []string
	NetworkLogPrint = func(text string) { messages = append(messages, text) }
	cases := []struct{ format, want string }{
		{"queue %0600d %s", "queue " + strings.Repeat("0", 505)},
		{"queue %0505d", "queue " + strings.Repeat("0", 503) + "37"},
		{"queue %04d %s", "queue 0037 ready"},
		{"queue %04d %s %d|%d|%d|%d|%d|%d|%s", "queue 0037 ready 2|3|4|5|6|7|spill"},
	}
	for _, tc := range cases {
		format := CString(tc.format)
		cfixture.NetworkLogPrintf413D30((*byte)(unsafe.Pointer(format)))
		StrFree(format)
	}
	if len(messages) != len(cases) {
		t.Fatalf("network log = %q", messages)
	}
	for i, tc := range cases {
		if messages[i] != tc.want {
			t.Errorf("format %q logged %q, want %q", tc.format, messages[i], tc.want)
		}
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
	const mixed = "[   ab|0x000012|+2.50|"
	func() {
		defer nox_fs_close(handle)
		for _, test := range []struct{ width, written int }{{4, 4}, {3000, 2047}} {
			if n := cfixture.FilePrintf(unsafe.Pointer(handle), test.width); n != test.written {
				t.Fatalf("width %d wrote %d bytes, want %d", test.width, n, test.written)
			}
		}
		if n := cfixture.FileMixedPrintf(unsafe.Pointer(handle)); n != len(mixed) {
			t.Fatalf("mixed format wrote %d bytes, want %d", n, len(mixed))
		}
	}()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "0007"+strings.Repeat("0", 2047)+mixed {
		t.Fatalf("unexpected file contents of %d bytes", len(content))
	}
}

func TestLegacyVPrintfPreservesCallerCursor(t *testing.T) {
	if !cfixture.VPrintfPreservesCursor() {
		t.Fatal("bounded or unbounded printf changed its input cursor, output, or guard bytes")
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

func TestLegacyFixedWideDoubleMatchesVariadic(t *testing.T) {
	format, release := CWString("%.62f|%d")
	defer release()
	for _, bits := range []uint64{
		math.Float64bits(-2.5), math.Float64bits(math.MaxFloat64),
		0x8000000000000000, 1, 0xfff8000000000000, 0x7ff0000000000001,
	} {
		var expected, actual [64]uint16
		for i := range expected {
			expected[i], actual[i] = 0xa5a5, 0xa5a5
		}
		want := cfixture.WideDoublePrintf(&expected[1], (*uint16)(unsafe.Pointer(format)), bits, -17)
		used := formatFixedWide((*wchar2_t)(unsafe.Pointer(&actual[1])), format,
			formatValue{doubleBits: bits}, formatValue{number: -17})
		if int(used) != want || actual != expected {
			t.Fatalf("double %016x fixed output (%d, %x) differs from variadic output (%d, %x)", bits, used, actual, want, expected)
		}
	}
}

func TestLegacyVariadicWordsAcrossRegistersAndStack(t *testing.T) {
	if !cfixture.VariadicRegistersAndStack() {
		t.Fatal("native format changed a mixed argument, terminator, or guard")
	}
}

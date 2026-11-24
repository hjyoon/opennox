package legacy

import (
	"math"
	"strings"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
)

func TestScriptCallbackQualifyExport542BF0PreservesSignedDwords(t *testing.T) {
	old := scriptCallbackQualifyCall542BF0
	t.Cleanup(func() { scriptCallbackQualifyCall542BF0 = old })

	var got [3]int32
	scriptCallbackQualifyCall542BF0 = func(a1, a2, a3 int32) {
		got = [3]int32{a1, a2, a3}
	}
	scriptCallbackQualifyExportCall542BF0(math.MinInt32, -17, math.MaxInt32)
	if want := [3]int32{math.MinInt32, -17, math.MaxInt32}; got != want {
		t.Fatalf("export arguments = %v, want %v", got, want)
	}
}

func TestScriptCallbackQualifyCHelpers542BF0PreserveNativeStringPointersAndBounds(t *testing.T) {
	got, callbackPointer := scriptCallbackQualifyNameExportCall5435C0(
		"Callback", math.MinInt32, -1, math.MaxInt32,
	)
	if want := "Callback%-2147483648%-1%2147483647"; got != want {
		t.Fatalf("C callback helper = %q, want %q", got, want)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && callbackPointer <= math.MaxUint32 {
		t.Fatalf("C callback input pointer = %#x, want native address above 4 GiB", callbackPointer)
	}
	if got, _ := scriptCallbackQualifyNameExportCall5435C0(strings.Repeat("a", 121), 0, 0, 0); len(got) != 127 {
		t.Fatalf("C 127-byte callback = %q (len %d), want preserved", got, len(got))
	}
	if got, _ := scriptCallbackQualifyNameExportCall5435C0(strings.Repeat("a", 122), 0, 0, 0); got != scriptCallbackError542BF0 {
		t.Fatalf("C 128-byte callback = %q, want %q", got, scriptCallbackError542BF0)
	}

	got, objectPointer := scriptObjectQualifyNameExportCall543620("Object", math.MinInt32)
	if want := "Object%-2147483648"; got != want {
		t.Fatalf("C object helper = %q, want %q", got, want)
	}
	if unsafe.Sizeof(uintptr(0)) == 8 && objectPointer <= math.MaxUint32 {
		t.Fatalf("C object input pointer = %#x, want native address above 4 GiB", objectPointer)
	}
	if got, _ := scriptObjectQualifyNameExportCall543620(strings.Repeat("b", 73), 0); len(got) != 75 {
		t.Fatalf("C 75-byte object = %q (len %d), want preserved", got, len(got))
	}
	if got, _ := scriptObjectQualifyNameExportCall543620(strings.Repeat("b", 74), 0); got != scriptCallbackError542BF0 {
		t.Fatalf("C 76-byte object = %q, want %q", got, scriptCallbackError542BF0)
	}
}

func TestScriptCallbackQualifyCHelpersShareScratchBuffer(t *testing.T) {
	buf := memmap.Slice(0x5D4594, 2489164)[:256]
	saved := [256]byte(buf)
	t.Cleanup(func() { copy(buf, saved[:]) })
	callback := sub_5435C0(nil, 1, -2, 3)
	if got := GoString(callback); got != "(null)%1%-2%3" {
		t.Fatalf("null callback formats as %q", got)
	}
	object := sub_543620(nil, -1)
	if callback != object || unsafe.Pointer(object) != unsafe.Pointer(&buf[0]) || GoString(callback) != "(null)%-1" {
		t.Fatal("object and callback qualifiers did not share the native output buffer")
	}
	name := strings.Repeat("a", 122)
	scriptCallbackQualifyNameExportCall5435C0(name, 0, 0, 0)
	start := len(scriptCallbackError542BF0) + 1
	if got, want := string(buf[start:129]), name[start:]+"%0%0%0\x00"; got != want {
		t.Fatalf("scratch after the error terminator = %q, want %q", got, want)
	}
}

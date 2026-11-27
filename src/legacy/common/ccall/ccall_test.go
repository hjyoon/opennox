package ccall_test

import (
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
	"github.com/opennox/opennox/v1/legacy/common/ccall/internal/testcall"
)

func TestSixScalarArguments(t *testing.T) {
	words, w := testcall.Words()
	if got := ccall.CallUPtrUPtr6(words, w[0], w[1], w[2], w[3], w[4], w[5]); got != w[5] {
		t.Fatalf("callback with six uintptr arguments returned %#x, want %#x", got, w[5])
	}
	ints, i := testcall.Ints()
	if got := ccall.CallIntInt6(ints, i[0], i[1], i[2], i[3], i[4], i[5]); got != i[5] {
		t.Fatalf("callback with six int arguments returned %d, want %d", got, i[5])
	}
}

func TestSixPinnedPointers(t *testing.T) {
	data := new([6]byte)
	var pin runtime.Pinner
	pin.Pin(data)
	defer pin.Unpin()

	got := ccall.CallPtrPtr6(testcall.Pointers(),
		unsafe.Pointer(&data[0]), unsafe.Pointer(&data[1]), unsafe.Pointer(&data[2]),
		unsafe.Pointer(&data[3]), unsafe.Pointer(&data[4]), unsafe.Pointer(&data[5]))
	if want := unsafe.Pointer(&data[5]); got != want {
		t.Fatalf("callback returned %p, want %p", got, want)
	}
	if want := [6]byte{1, 2, 3, 4, 5, 6}; *data != want {
		t.Fatalf("callback wrote %v, want %v", *data, want)
	}
}

package legacy

import (
	"bytes"
	"math"
	"testing"
	"unsafe"
)

func TestMemfileReaderNativeCursor(t *testing.T) {
	input := []byte{
		0xf4, 0x34, 0x12, 0, 0, 0, 0x80, 0xff,
		9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 20, 21,
	}
	file, free := thingFixtureMemfile411540(input, 0)
	defer free()

	if got := nox_memfile_read_i8(file); got != -12 {
		t.Fatalf("signed byte = %d", got)
	}
	if got := nox_memfile_read_u16(file); got != 0x1234 {
		t.Fatalf("little-endian word = %#x", got)
	}
	if got := nox_memfile_read_i32(file); got != math.MinInt32 {
		t.Fatalf("signed dword = %d", got)
	}

	aligned := [3]byte{}
	if got := memfileReadAligned(aligned[:], file); got != 1 ||
		aligned != [3]byte{9, 10, 11} || thingFixtureOffset411540(file) != 16 {
		t.Fatalf("aligned record = %v, return %d, cursor %d", aligned, got, thingFixtureOffset411540(file))
	}

	last := [6]byte{0, 0, 0, 0, 0, 0x99}
	if got := nox_memfile_read(unsafe.Pointer(&last[0]), 2, 3, file); got != 2 ||
		!bytes.Equal(last[:5], input[16:]) || last[5] != 0x99 || thingFixtureOffset411540(file) != len(input) {
		t.Fatalf("partial record = %v, return %d, cursor %d", last, got, thingFixtureOffset411540(file))
	}
	if got := nox_memfile_read_u32(file); got != 0 || thingFixtureOffset411540(file) != len(input) {
		t.Fatalf("scalar at EOF = %d, cursor %d", got, thingFixtureOffset411540(file))
	}
	nox_memfile_skip(file, -100)
	if got := thingFixtureOffset411540(file); got != 0 {
		t.Fatalf("rewind before start left cursor at %d", got)
	}
}

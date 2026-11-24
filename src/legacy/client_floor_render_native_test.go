package legacy

import (
	"bytes"
	"testing"
)

func TestFloorSpansPreserveOffsetsAcrossBufferWrap(t *testing.T) {
	buf := make([]byte, 5)
	copyFloorSpan(buf, -2, []byte{1, 2, 3, 4, 5, 6, 7, 8})
	if want := []byte{8, 4, 5, 6, 7}; !bytes.Equal(buf, want) {
		t.Fatalf("wrapped copy = %v, want %v", buf, want)
	}
	fillFloorSpan(buf, 9, 0x44332211, 7)
	if want := []byte{0x33, 0x33, 0x44, 0x11, 0x22}; !bytes.Equal(buf, want) {
		t.Fatalf("wrapped fill = %v, want %v", buf, want)
	}
}

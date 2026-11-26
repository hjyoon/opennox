package legacy

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
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

func TestFloorBufferUsesNativeRangeUntilRelease(t *testing.T) {
	buffer, free := alloc.MakeHigh([]byte{1, 2, 3, 4}, 4)
	defer free()
	start, end := floorTileBufferStart, floorTileBufferEnd
	defer func() {
		Set_nox_video_tileBuf_ptr_3798796(start)
		Set_nox_video_tileBuf_end_3798844(end)
	}()

	p := unsafe.Pointer(&buffer[0])
	Set_nox_video_tileBuf_ptr_3798796(p)
	Set_nox_video_tileBuf_end_3798844(unsafe.Add(p, len(buffer)))
	if got := floorTileBuffer(); !bytes.Equal(got, buffer) || &got[0] != &buffer[0] {
		t.Fatalf("floor buffer view = %v, want a view of %v", got, buffer)
	}
	Set_nox_video_tileBuf_ptr_3798796(nil)
	Set_nox_video_tileBuf_end_3798844(nil)
	if got := floorTileBuffer(); got != nil {
		t.Fatalf("released floor buffer view = %v", got)
	}
}

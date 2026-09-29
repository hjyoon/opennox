//go:build !server

package legacy

import (
	"image"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
)

func TestWallSort476080UsesNativePlayerPointerAndPosition(t *testing.T) {
	playerSlot := memmap.PtrPtr(0x852978, 8)
	oldPlayer := *playerSlot
	t.Cleanup(func() { *playerSlot = oldPlayer })

	wall := [7]byte{0, 0, 0, 0, 0, 5, 6}
	*playerSlot = nil
	if got := Sub_476080(unsafe.Pointer(&wall[0])); got != 149 {
		t.Fatalf("sort key without player = %d, want 149", got)
	}

	player := &client.Drawable{PosVec: image.Pt(100, 100)}
	var pin runtime.Pinner
	pin.Pin(player)
	defer pin.Unpin()
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(player)) <= math.MaxUint32 {
		t.Fatalf("drawable pointer = %p, want native address above 4 GiB", player)
	}

	*playerSlot = unsafe.Pointer(player)
	if got := Sub_476080(unsafe.Pointer(&wall[0])); got != 160 {
		t.Fatalf("sort key at (100,100) = %d, want 160", got)
	}

	player.PosVec = image.Pt(200, 200)
	if got := Sub_476080(unsafe.Pointer(&wall[0])); got != 138 {
		t.Fatalf("sort key at (200,200) = %d, want 138", got)
	}
	runtime.KeepAlive(player)
}

//go:build !server

package opennox

import (
	"image"
	"math"
	"runtime"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/common/memmap"
)

func TestDrawableSortKey4761B0(t *testing.T) {
	dr := &client.Drawable{PosVec: image.Pt(10, 20)}
	player := &client.Drawable{PosVec: image.Pt(30, 50)}
	const dir = byte(251)
	dr.Field_74_4 = dir
	dirOff := uintptr(dir) * 8
	dirX := memmap.PtrInt32(0x587000, 196184+dirOff)
	dirY := memmap.PtrInt32(0x587000, 196188+dirOff)
	oldX, oldY := *dirX, *dirY
	t.Cleanup(func() {
		*dirX, *dirY = oldX, oldY
	})

	tests := []struct {
		name   string
		player *client.Drawable
		x      int32
		y      int32
		want   int32
	}{
		{name: "no player truncates negative half toward zero", x: 4, y: -5, want: 18},
		{name: "nonnegative cross selects descending endpoint", player: player, x: 4, y: -6, want: 14},
		{name: "negative cross selects ascending endpoint", player: &client.Drawable{PosVec: image.Pt(100, 50)}, x: 4, y: 6, want: 26},
		{name: "negative direction reverses cross sign", player: player, x: -4, y: 6, want: 20},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			*dirX, *dirY = test.x, test.y
			if got := drawableSortKey4761B0(dr, test.player); got != test.want {
				t.Fatalf("sort key = %d, want %d", got, test.want)
			}
		})
	}
}

func TestDrawableSort476160PreservesNativePointers(t *testing.T) {
	dr := &client.Drawable{
		PosVec:     image.Pt(10, 20),
		ObjClass:   object.Class(0x80),
		Field_74_4: 250,
	}
	player := &client.Drawable{PosVec: image.Pt(30, 50)}
	other := &client.Drawable{PosVec: image.Pt(10, 15)}
	var pin runtime.Pinner
	defer pin.Unpin()
	for _, ptr := range []unsafe.Pointer{unsafe.Pointer(dr), unsafe.Pointer(player), unsafe.Pointer(other)} {
		pin.Pin(ptr)
		if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(ptr) <= math.MaxUint32 {
			t.Fatalf("drawable pointer = %p, want native address above 4 GiB", ptr)
		}
	}
	dirOff := uintptr(dr.Field_74_4) * 8
	dirX := memmap.PtrInt32(0x587000, 196184+dirOff)
	dirY := memmap.PtrInt32(0x587000, 196188+dirOff)
	oldX, oldY := *dirX, *dirY
	*dirX, *dirY = 4, -6
	t.Cleanup(func() {
		*dirX, *dirY = oldX, oldY
	})

	playerSlot := memmap.PtrPtr(0x852978, 8)
	oldPlayer := *playerSlot
	*playerSlot = unsafe.Pointer(player)
	t.Cleanup(func() {
		*playerSlot = oldPlayer
	})

	c := new(Client)
	if !c.sub_476160(dr, other) {
		t.Fatal("special key 14 should sort before ordinary key 15")
	}
	if c.sub_476160(other, dr) {
		t.Fatal("ordinary key 15 should not sort before special key 14")
	}

	signedZ := &client.Drawable{PosVec: image.Pt(0, 20), ZVal: 0xffff}
	if got := drawableSortKey476160(signedZ, player); got != 19 {
		t.Fatalf("signed Z sort key = %d, want 19", got)
	}
	runtime.KeepAlive(dr)
	runtime.KeepAlive(player)
	runtime.KeepAlive(other)
}

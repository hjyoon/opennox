package opennox

import (
	"encoding/binary"
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func spikeFrameHooks518AE0() netSpriteUpdateHooks518AE0 {
	return netSpriteUpdateHooks518AE0{typeIndex: func(id string) int {
		if id == "Spike" {
			return 12
		}
		return 0
	}}
}

func TestSpikeFrame518AE0UsesOriginalHighFlagByte(t *testing.T) {
	// GAME.EXE 00518B5E..00518B6D: flags DWORD, SHR EAX,24, NOT AL,
	// AND AL,1. Literal masks are independent of the port's flag names.
	noise := []uint32{0, 0xffffffff &^ 0x01000000}
	for bit := uint(0); bit < 32; bit++ {
		if bit != 24 {
			noise = append(noise, uint32(1)<<bit)
		}
	}
	for _, other := range noise {
		for _, enabled := range []bool{false, true} {
			flags, want := other, byte(1)
			if enabled {
				flags |= 0x01000000
				want = 0
			}
			t.Run(fmt.Sprintf("flags_%08x", flags), func(t *testing.T) {
				obj := &server.Object{TypeInd: 12, ObjClass: object.ClassImmobile, ObjFlags: object.Flags(flags)}
				state, ok := netSpriteUpdateStateNative518AE0(obj, spikeFrameHooks518AE0())
				if !ok || state.opcode != netmsg.MSG_DRAW_FRAME || state.value != want || state.direct {
					t.Fatalf("flags=%08x state=%+v selected=%t; want ordinary draw frame %d", flags, state, ok, want)
				}
			})
		}
	}
}

func TestSpikeOnOff518AE0ReachesClientSlaveFrame(t *testing.T) {
	obj, freeObject := alloc.New(server.Object{})
	*obj = server.Object{TypeInd: 12, ObjClass: object.ClassImmobile, ObjFlags: object.FlagEquipped}
	dr, freeDrawable := alloc.New(client.Drawable{})
	*dr = client.Drawable{ObjClass: object.ClassImmobile, AnimFrameSlave: 7}
	t.Cleanup(freeObject)
	t.Cleanup(freeDrawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(obj)) <= math.MaxUint32 || uintptr(unsafe.Pointer(dr)) <= math.MaxUint32) {
		t.Fatal("native fixture pointers must be above 4 GiB")
	}
	const code = uint16(0x9234)
	hooks := objectActivationHooks48EA70{connected: func() bool { return true }, byNetCode: func(got uint16) *client.Drawable {
		if got != code {
			t.Fatalf("decoded code=%04x, want %04x", got, code)
		}
		return dr
	}}
	base := &server.Server{}
	for i, enabled := range []bool{false, true, true, false, true, false} {
		previous := dr.AnimFrameSlave
		obj.SetOnOff(enabled)
		base.SwitchUpdate53B320(obj, server.SwitchUpdateRuntime53B320{})
		if obj.Flags().Has(object.FlagNoCollide) == enabled || !obj.Flags().Has(object.FlagEquipped) || obj.Field38 != math.MaxUint32 {
			t.Fatalf("step%d collision/equipment/sync state=%08x/%08x", i, obj.Flags(), obj.Field38)
		}
		state, ok := netSpriteUpdateStateNative518AE0(obj, spikeFrameHooks518AE0())
		want := byte(1)
		if enabled {
			want = 0
		}
		if !ok || state.value != want {
			t.Fatalf("step%d server draw frame=%d, want %d", i, state.value, want)
		}
		packet := []byte{byte(state.opcode), 0, 0, state.value}
		binary.LittleEndian.PutUint16(packet[1:], code)
		if got := handleObjectDrawFrameNative48EA70(packet, hooks); got != 4 || dr.AnimFrameSlave != uint32(want) || dr.Field_78 != previous {
			t.Fatalf("step%d client result=%d frame=%d previous=%d; want 4/%d/%d", i, got, dr.AnimFrameSlave, dr.Field_78, want, previous)
		}
	}
}

package opennox

import (
	"math"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
)

func highAddressObjectActivationDrawable48EA70(t *testing.T) *client.Drawable {
	t.Helper()
	dr := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(math.MaxUint32) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	return dr
}

func TestDecodeObjectActivationStates48EA70(t *testing.T) {
	activation := []byte{byte(netmsg.MSG_ENABLE_OBJECT), 0x34, 0x92, 0xaa}
	state, ok := decodeObjectActivationState48EA70(activation)
	if !ok || state.Code != 0x9234 {
		t.Fatalf("activation state = %+v, ok=%t", state, ok)
	}
	for n := 0; n < objectActivationPacketSize48EA70; n++ {
		if _, ok := decodeObjectActivationState48EA70(activation[:n]); ok {
			t.Fatalf("%d-byte activation packet was accepted", n)
		}
	}

	framePacket := []byte{byte(netmsg.MSG_DRAW_FRAME), 0x78, 0x56, 0xab, 0xcc}
	frame, ok := decodeObjectDrawFrameState48EA70(framePacket)
	if !ok || frame.Code != 0x5678 || frame.Frame != 0xab {
		t.Fatalf("draw-frame state = %+v, ok=%t", frame, ok)
	}
	for n := 0; n < objectDrawFramePacketSize48EA70; n++ {
		if _, ok := decodeObjectDrawFrameState48EA70(framePacket[:n]); ok {
			t.Fatalf("%d-byte draw-frame packet was accepted", n)
		}
	}
}

func TestHandleObjectActivationNative48EA70PreservesHighAddressAndNamespace(t *testing.T) {
	dr := highAddressObjectActivationDrawable48EA70(t)
	draw := unsafe.Pointer(dr)
	dr.ObjClass = object.ClassReadable
	dr.ObjFlags = object.FlagActive
	dr.DrawFuncPtr = draw
	var codes []uint16
	hooks := objectActivationHooks48EA70{
		connected: func() bool { return true },
		byNetCode: func(code uint16) *client.Drawable {
			codes = append(codes, code)
			return dr
		},
	}

	enable := []byte{byte(netmsg.MSG_ENABLE_OBJECT), 0x34, 0x92}
	if got := handleObjectActivationNative48EA70(enable, true, hooks); got != len(enable) {
		t.Fatalf("enable consumed bytes = %d, want %d", got, len(enable))
	}
	if !dr.ObjFlags.Has(object.FlagEnabled) || dr.DrawFuncPtr != draw {
		t.Fatalf("enabled object = flags %#x, draw %p; want enabled and draw %p", dr.ObjFlags, dr.DrawFuncPtr, draw)
	}

	disable := []byte{byte(netmsg.MSG_DISABLE_OBJECT), 0x34, 0x92}
	if got := handleObjectActivationNative48EA70(disable, false, hooks); got != len(disable) {
		t.Fatalf("disable consumed bytes = %d, want %d", got, len(disable))
	}
	if dr.ObjFlags.Has(object.FlagEnabled) || dr.DrawFuncPtr != nil {
		t.Fatalf("disabled object = flags %#x, draw %p; want disabled and nil draw", dr.ObjFlags, dr.DrawFuncPtr)
	}
	if want := []uint16{0x9234, 0x9234}; !reflect.DeepEqual(codes, want) {
		t.Fatalf("lookup codes = %#v, want %#v", codes, want)
	}
}

func TestHandleObjectActivationNative48EA70PreservesOtherDrawCallbacks(t *testing.T) {
	dr := highAddressObjectActivationDrawable48EA70(t)
	draw := unsafe.Pointer(dr)
	dr.ObjClass = object.ClassMonster
	dr.ObjFlags = object.FlagEnabled
	dr.DrawFuncPtr = draw
	hooks := objectActivationHooks48EA70{
		connected: func() bool { return true },
		byNetCode: func(uint16) *client.Drawable { return dr },
	}
	if got := handleObjectActivationNative48EA70([]byte{byte(netmsg.MSG_DISABLE_OBJECT), 1, 0}, false, hooks); got != objectActivationPacketSize48EA70 {
		t.Fatalf("consumed bytes = %d, want %d", got, objectActivationPacketSize48EA70)
	}
	if dr.ObjFlags.Has(object.FlagEnabled) || dr.DrawFuncPtr != draw {
		t.Fatalf("disabled ordinary object = flags %#x, draw %p; want callback %p preserved", dr.ObjFlags, dr.DrawFuncPtr, draw)
	}
}

func TestHandleObjectDrawFrameNative48EA70PreservesHighAddress(t *testing.T) {
	dr := highAddressObjectActivationDrawable48EA70(t)
	dr.AnimFrameSlave = 7
	var code uint16
	hooks := objectActivationHooks48EA70{
		connected: func() bool { return true },
		byNetCode: func(got uint16) *client.Drawable {
			code = got
			return dr
		},
	}
	packet := []byte{byte(netmsg.MSG_DRAW_FRAME), 0x34, 0x12, 0xab}
	before := append([]byte(nil), packet...)
	if got := handleObjectDrawFrameNative48EA70(packet, hooks); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if code != 0x1234 || dr.Field_78 != 7 || dr.AnimFrameSlave != 0xab {
		t.Fatalf("draw-frame result = code %#x, previous %d, current %#x", code, dr.Field_78, dr.AnimFrameSlave)
	}
	if !reflect.DeepEqual(packet, before) {
		t.Fatalf("packet mutated: got %x, want %x", packet, before)
	}
}

func TestHandleObjectActivationNative48EA70Guards(t *testing.T) {
	lookups := 0
	hooks := objectActivationHooks48EA70{
		connected: func() bool { return false },
		byNetCode: func(uint16) *client.Drawable {
			lookups++
			return nil
		},
	}
	if got := handleObjectActivationNative48EA70([]byte{byte(netmsg.MSG_ENABLE_OBJECT), 1, 0}, true, hooks); got != objectActivationPacketSize48EA70 {
		t.Fatalf("disconnected activation consumed %d bytes", got)
	}
	if got := handleObjectDrawFrameNative48EA70([]byte{byte(netmsg.MSG_DRAW_FRAME), 1, 0, 9}, hooks); got != objectDrawFramePacketSize48EA70 {
		t.Fatalf("disconnected draw-frame consumed %d bytes", got)
	}
	if lookups != 0 {
		t.Fatalf("disconnected handlers performed %d lookups", lookups)
	}
	if got := handleObjectActivationNative48EA70([]byte{byte(netmsg.MSG_ENABLE_OBJECT), 1}, true, objectActivationHooks48EA70{}); got != -1 {
		t.Fatalf("short activation result = %d, want -1", got)
	}
	if got := handleObjectDrawFrameNative48EA70([]byte{byte(netmsg.MSG_DRAW_FRAME), 1, 0}, objectActivationHooks48EA70{}); got != -1 {
		t.Fatalf("short draw-frame result = %d, want -1", got)
	}
}

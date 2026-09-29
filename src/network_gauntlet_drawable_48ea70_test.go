package opennox

import (
	"encoding/binary"
	"image"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client"
)

func TestHandleGauntletHideStaticDrawableNative48EA70HighAddress(t *testing.T) {
	dr := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	*(*byte)(unsafe.Pointer(&dr.Union)) = 0xa5
	data := []byte{byte(netmsg.MSG_GAUNTLET), gauntletHideStaticDrawable48EA70, 0x34, 0x92, 0xcc}
	before := append([]byte(nil), data...)
	var calls []string
	got, handled := handleGauntletDrawableNative48EA70(data, gauntletDrawableHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return true
		},
		lookupStatic: func(code uint16) *client.Drawable {
			calls = append(calls, "lookup")
			if code != 0x9234 {
				t.Fatalf("static lookup code = %#x, want 0x9234", code)
			}
			return dr
		},
	})
	if !handled || got != 4 {
		t.Fatalf("result = (%d, %t), want (4, true)", got, handled)
	}
	if first := *(*byte)(unsafe.Pointer(&dr.Union)); first != 0 {
		t.Fatalf("drawable union first byte = %#x, want 0", first)
	}
	if want := []string{"connected", "lookup"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
	if !reflect.DeepEqual(data, before) {
		t.Fatalf("packet mutated: got %x, want %x", data, before)
	}
}

func TestHandleGauntletGreenZapNative48EA70HighAddress(t *testing.T) {
	dr := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	data := make([]byte, 13)
	data[0] = byte(netmsg.MSG_GAUNTLET)
	data[1] = gauntletGreenZap48EA70
	binary.LittleEndian.PutUint16(data[2:4], 0x1234)
	binary.LittleEndian.PutUint16(data[4:6], 0xabcd)
	binary.LittleEndian.PutUint16(data[6:8], 0xffff)
	binary.LittleEndian.PutUint16(data[8:10], 0x8000)
	binary.LittleEndian.PutUint16(data[10:12], 0x4567)
	data[12] = 0xcc
	before := append([]byte(nil), data...)
	var calls []string
	got, handled := handleGauntletDrawableNative48EA70(data, gauntletDrawableHooks48EA70{
		typeID: func() int {
			calls = append(calls, "type")
			return 77
		},
		connected: func() bool {
			calls = append(calls, "connected")
			return true
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			calls = append(calls, "spawn")
			if typ != 77 || pos != image.Pt(0xffff, 0x8000) {
				t.Fatalf("spawn = type %d at %v, want type 77 at (65535,32768)", typ, pos)
			}
			return dr
		},
		decay: func(got *client.Drawable, lifetime int) {
			calls = append(calls, "decay")
			if got != dr || lifetime != 0x4567 {
				t.Fatalf("decay = (%p, %#x), want (%p, 0x4567)", got, lifetime, dr)
			}
		},
	})
	if !handled || got != 12 {
		t.Fatalf("result = (%d, %t), want (12, true)", got, handled)
	}
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&dr.Union)), 13)
	wantPayload := []byte{0, 0x67, 0x45, 0, 0, 0x34, 0x12, 0xcd, 0xab, 0xff, 0xff, 0x00, 0x80}
	if !reflect.DeepEqual(payload, wantPayload) {
		t.Fatalf("GreenZap payload = %x, want %x", payload, wantPayload)
	}
	if want := []string{"type", "connected", "spawn", "decay"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
	if !reflect.DeepEqual(data, before) {
		t.Fatalf("packet mutated: got %x, want %x", data, before)
	}
}

func TestHandleGauntletDrawableNative48EA70Gates(t *testing.T) {
	hide := []byte{byte(netmsg.MSG_GAUNTLET), gauntletHideStaticDrawable48EA70, 1, 0}
	for n := 0; n < len(hide); n++ {
		if got, handled := handleGauntletDrawableNative48EA70(hide[:n], gauntletDrawableHooks48EA70{}); !handled || got != -1 {
			t.Fatalf("%d-byte hide packet result = (%d, %t), want (-1, true)", n, got, handled)
		}
	}
	green := []byte{byte(netmsg.MSG_GAUNTLET), gauntletGreenZap48EA70, 1, 0, 2, 0, 3, 0, 4, 0, 5, 0}
	for n := 0; n < len(green); n++ {
		if got, handled := handleGauntletDrawableNative48EA70(green[:n], gauntletDrawableHooks48EA70{}); !handled || got != -1 {
			t.Fatalf("%d-byte GreenZap packet result = (%d, %t), want (-1, true)", n, got, handled)
		}
	}

	called := false
	if got, handled := handleGauntletDrawableNative48EA70(hide, gauntletDrawableHooks48EA70{
		connected: func() bool { return false },
		lookupStatic: func(uint16) *client.Drawable {
			called = true
			return nil
		},
	}); !handled || got != 4 || called {
		t.Fatalf("disconnected hide result/callback = (%d, %t)/%t, want (4, true)/false", got, handled, called)
	}

	var calls []string
	if got, handled := handleGauntletDrawableNative48EA70(green, gauntletDrawableHooks48EA70{
		typeID: func() int {
			calls = append(calls, "type")
			return 1
		},
		connected: func() bool {
			calls = append(calls, "connected")
			return false
		},
		spawn: func(int, image.Point) *client.Drawable {
			calls = append(calls, "spawn")
			return nil
		},
	}); !handled || got != 12 || !reflect.DeepEqual(calls, []string{"type", "connected"}) {
		t.Fatalf("disconnected GreenZap result/calls = (%d, %t)/%v", got, handled, calls)
	}

	called = false
	if got, handled := handleGauntletDrawableNative48EA70([]byte{byte(netmsg.MSG_GAUNTLET), 0x14}, gauntletDrawableHooks48EA70{
		connected: func() bool {
			called = true
			return true
		},
	}); handled || got != 0 || called {
		t.Fatalf("unsupported subcommand result/callback = (%d, %t)/%t, want (0, false)/false", got, handled, called)
	}
}

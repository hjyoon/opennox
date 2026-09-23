package opennox

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/client"
)

func highAddressFixedObjectDrawable48EA70(t *testing.T) *client.Drawable {
	t.Helper()
	dr := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(math.MaxUint32) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	return dr
}

func fixedObjectStateHooksFor48EA70(t *testing.T, dr *client.Drawable, wantCode uint16, frame uint32) fixedObjectStateHooks48EA70 {
	t.Helper()
	return fixedObjectStateHooks48EA70{
		connected: func() bool { return true },
		byNetCode: func(code uint16) *client.Drawable {
			if code != wantCode {
				t.Fatalf("lookup code = %#x, want %#x", code, wantCode)
			}
			return dr
		},
		frame: func() uint32 { return frame },
	}
}

func TestHandleDoorAngleNative48EA70PreservesHighAddressDrawable(t *testing.T) {
	dr := highAddressFixedObjectDrawable48EA70(t)
	packet := []byte{byte(netmsg.MSG_DOOR_ANGLE), 0x34, 0x92, 0xa7}
	if got := handleFixedObjectStateNative48EA70(netmsg.MSG_DOOR_ANGLE, packet,
		fixedObjectStateHooksFor48EA70(t, dr, 0x9234, 0)); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if dr.Field_74_4 != 0xa7 {
		t.Fatalf("door angle = %#x, want 0xa7", dr.Field_74_4)
	}
}

func TestHandleObeliskChargeNative48EA70PreservesHighAddressDrawable(t *testing.T) {
	dr := highAddressFixedObjectDrawable48EA70(t)
	dr.ObjClass = object.ClassImmobile
	dr.AnimFrameSlave = 8
	packet := []byte{byte(netmsg.MSG_OBELISK_CHARGE), 0x34, 0x12, 50}
	if got := handleFixedObjectStateNative48EA70(netmsg.MSG_OBELISK_CHARGE, packet,
		fixedObjectStateHooksFor48EA70(t, dr, 0x1234, 0)); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if !dr.ObjFlags.Has(object.FlagActive) {
		t.Fatal("obelisk drawable was not activated")
	}
	if dr.LightIntensity != 63 || dr.LightIntensityRad == 0 || dr.LightIntensityU16 == 0 {
		t.Fatalf("obelisk light = intensity %v, radius %d, fixed %d", dr.LightIntensity, dr.LightIntensityRad, dr.LightIntensityU16)
	}
	if dr.AnimFrameSlave != 7 {
		t.Fatalf("obelisk frame = %d, want corrected frame 7", dr.AnimFrameSlave)
	}
}

func TestHandlePentagramActivateNative48EA70PreservesHighAddressDrawable(t *testing.T) {
	dr := highAddressFixedObjectDrawable48EA70(t)
	dr.ObjClass = object.ClassTrigger
	hooks := fixedObjectStateHooksFor48EA70(t, dr, 0x9234, 0x89abcdef)

	active := []byte{byte(netmsg.MSG_PENTAGRAM_ACTIVATE), 0x34, 0x92, 1}
	if got := handleFixedObjectStateNative48EA70(netmsg.MSG_PENTAGRAM_ACTIVATE, active, hooks); got != len(active) {
		t.Fatalf("active consumed bytes = %d, want %d", got, len(active))
	}
	if !dr.ObjClass.Has(object.ClassLight) || dr.LightIntensity != float32(41.958) ||
		dr.AnimFrameSlave != 1 || dr.Field_72 != 0x89abcdef {
		t.Fatalf("active pentagram state = class %#x, intensity %v, frame %d, tick %#x",
			dr.ObjClass, dr.LightIntensity, dr.AnimFrameSlave, dr.Field_72)
	}

	inactive := []byte{byte(netmsg.MSG_PENTAGRAM_ACTIVATE), 0x34, 0x92, 0}
	if got := handleFixedObjectStateNative48EA70(netmsg.MSG_PENTAGRAM_ACTIVATE, inactive, hooks); got != len(inactive) {
		t.Fatalf("inactive consumed bytes = %d, want %d", got, len(inactive))
	}
	if dr.ObjClass.Has(object.ClassLight) || dr.LightIntensity != 0 || dr.AnimFrameSlave != 0 {
		t.Fatalf("inactive pentagram state = class %#x, intensity %v, frame %d",
			dr.ObjClass, dr.LightIntensity, dr.AnimFrameSlave)
	}
}

func TestHandleFixedObjectStateNative48EA70Guards(t *testing.T) {
	for n := 0; n < 4; n++ {
		if got := handleFixedObjectStateNative48EA70(netmsg.MSG_OBELISK_CHARGE, make([]byte, n), fixedObjectStateHooks48EA70{}); got != -1 {
			t.Fatalf("short packet of %d bytes consumed %d, want -1", n, got)
		}
	}
	lookups := 0
	hooks := fixedObjectStateHooks48EA70{
		connected: func() bool { return false },
		byNetCode: func(uint16) *client.Drawable { lookups++; return nil },
	}
	if got := handleFixedObjectStateNative48EA70(netmsg.MSG_OBELISK_CHARGE,
		[]byte{byte(netmsg.MSG_OBELISK_CHARGE), 1, 0, 1}, hooks); got != 4 || lookups != 0 {
		t.Fatalf("disconnected result = %d with %d lookups, want 4 with none", got, lookups)
	}
	hooks.connected = func() bool { return true }
	if got := handleFixedObjectStateNative48EA70(netmsg.MSG_OBELISK_CHARGE,
		[]byte{byte(netmsg.MSG_OBELISK_CHARGE), 1, 0, 1}, hooks); got != 4 || lookups != 1 {
		t.Fatalf("missing drawable result = %d with %d lookups, want 4 with one", got, lookups)
	}
}

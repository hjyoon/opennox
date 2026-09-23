package opennox

import (
	"encoding/binary"
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client"
)

func highAddressObjectLightDrawable48EA70(t *testing.T) *client.Drawable {
	t.Helper()
	dr := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(math.MaxUint32) {
		t.Skipf("allocator returned a low address: %p", dr)
	}
	return dr
}

func objectLightHooksFor48EA70(t *testing.T, dr *client.Drawable, wantCode uint16) objectLightHooks48EA70 {
	t.Helper()
	return objectLightHooks48EA70{
		connected: func() bool { return true },
		byNetCode: func(code uint16) *client.Drawable {
			if code != wantCode {
				t.Fatalf("lookup code = %#x, want %#x", code, wantCode)
			}
			return dr
		},
	}
}

func TestHandleObjectLightColorNative48EA70PreservesHighAddressDrawable(t *testing.T) {
	dr := highAddressObjectLightDrawable48EA70(t)
	packet := []byte{byte(netmsg.MSG_REPORT_LIGHT_COLOR), 0x34, 0x92, 17, 123, 251}
	if got := handleObjectLightNative48EA70(netmsg.MSG_REPORT_LIGHT_COLOR, packet,
		objectLightHooksFor48EA70(t, dr, 0x9234)); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if dr.LightFlags != 2 || dr.LightColor.R != 17 || dr.LightColor.G != 123 || dr.LightColor.B != 251 {
		t.Fatalf("light state = flags %d, color %+v", dr.LightFlags, dr.LightColor)
	}
}

func TestHandleObjectLightIntensityNative48EA70PreservesHighAddressDrawable(t *testing.T) {
	dr := highAddressObjectLightDrawable48EA70(t)
	packet := make([]byte, 7)
	packet[0] = byte(netmsg.MSG_REPORT_LIGHT_INTENSITY)
	binary.LittleEndian.PutUint16(packet[1:3], 0x1234)
	binary.LittleEndian.PutUint32(packet[3:7], math.Float32bits(75.5))
	if got := handleObjectLightNative48EA70(netmsg.MSG_REPORT_LIGHT_INTENSITY, packet,
		objectLightHooksFor48EA70(t, dr, 0x1234)); got != len(packet) {
		t.Fatalf("consumed bytes = %d, want %d", got, len(packet))
	}
	if dr.LightIntensity != 63 || dr.LightIntensityRad == 0 || dr.LightIntensityU16 != 63*0x10000 {
		t.Fatalf("light intensity state = %v, radius %d, fixed %d",
			dr.LightIntensity, dr.LightIntensityRad, dr.LightIntensityU16)
	}
}

func TestHandleObjectLightNative48EA70Guards(t *testing.T) {
	for _, tc := range []struct {
		op   netmsg.Op
		size int
	}{
		{netmsg.MSG_REPORT_LIGHT_COLOR, 6},
		{netmsg.MSG_REPORT_LIGHT_INTENSITY, 7},
	} {
		for n := 0; n < tc.size; n++ {
			if got := handleObjectLightNative48EA70(tc.op, make([]byte, n), objectLightHooks48EA70{}); got != -1 {
				t.Fatalf("op %v short packet of %d bytes consumed %d, want -1", tc.op, n, got)
			}
		}
	}
	lookups := 0
	hooks := objectLightHooks48EA70{
		connected: func() bool { return false },
		byNetCode: func(uint16) *client.Drawable { lookups++; return nil },
	}
	packet := []byte{byte(netmsg.MSG_REPORT_LIGHT_COLOR), 1, 0, 1, 2, 3}
	if got := handleObjectLightNative48EA70(netmsg.MSG_REPORT_LIGHT_COLOR, packet, hooks); got != 6 || lookups != 0 {
		t.Fatalf("disconnected result = %d with %d lookups, want 6 with none", got, lookups)
	}
	hooks.connected = func() bool { return true }
	if got := handleObjectLightNative48EA70(netmsg.MSG_REPORT_LIGHT_COLOR, packet, hooks); got != 6 || lookups != 1 {
		t.Fatalf("missing drawable result = %d with %d lookups, want 6 with one", got, lookups)
	}
	if got := handleObjectLightNative48EA70(netmsg.MSG_CODE2, packet, hooks); got != -1 {
		t.Fatalf("unsupported opcode result = %d, want -1", got)
	}
}

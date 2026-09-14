package opennox

import (
	"encoding/binary"
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
)

func TestDurationRayNative48EA70HighAddressLifecycle(t *testing.T) {
	from := &client.Drawable{PosVec: image.Pt(10, 20)}
	to := &client.Drawable{PosVec: image.Pt(51, 61)}
	ray := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ray)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", ray)
	}
	var slots [96]clientDurationRay48EA70
	var removed *client.Drawable
	hooks := durationRayHooks48EA70{
		connected: func() bool { return true },
		byCode: func(code uint16) *client.Drawable {
			switch code {
			case 0x1234:
				return from
			case 0x9234:
				return to
			default:
				t.Fatalf("unexpected net code %#x", code)
				return nil
			}
		},
		typeID: func(index int, name string) int {
			if index != 2 || name != "DynamicChainLightning" {
				t.Fatalf("ray type = %d/%q", index, name)
			}
			return 77
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ != 77 || pos != (image.Pt(30, 40)) {
				t.Fatalf("spawn = %d/%v", typ, pos)
			}
			return ray
		},
		remove: func(dr *client.Drawable) { removed = dr },
	}
	start := []byte{0x9e, 3, 5, 0x34, 0x12, 0x34, 0x92}
	if got := handleDurationRayNative48EA70(start, &slots, hooks); got != 7 {
		t.Fatalf("start consumed %d bytes", got)
	}
	if slots[0].drawable != ray || slots[0].source != 0x1234 || slots[0].target != 0x9234 || slots[0].kind != 3 {
		t.Fatalf("ray slot = %+v", slots[0])
	}
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&ray.Union)), 13)
	if payload[0] != 1 || binary.LittleEndian.Uint32(payload[1:]) != 5 ||
		binary.LittleEndian.Uint32(payload[5:]) != 0x1234 ||
		binary.LittleEndian.Uint32(payload[9:]) != 0x9234 {
		t.Fatalf("drawable ray payload = %x", payload)
	}
	// The server's stop packet reverses the two wire codes.
	stop := []byte{0x9e, 10, 0, 0x34, 0x92, 0x34, 0x12}
	if got := handleDurationRayNative48EA70(stop, &slots, hooks); got != 7 {
		t.Fatalf("stop consumed %d bytes", got)
	}
	if removed != ray || slots[0].drawable != nil {
		t.Fatalf("ray removal = %p, slot = %+v", removed, slots[0])
	}
}

func TestDurationRayNative48EA70PacketGuards(t *testing.T) {
	var slots [96]clientDurationRay48EA70
	for n := 0; n < 7; n++ {
		if got := handleDurationRayNative48EA70(make([]byte, n), &slots, durationRayHooks48EA70{}); got != -1 {
			t.Fatalf("short packet %d consumed %d bytes", n, got)
		}
	}
	for _, kind := range []byte{0, 15, 255} {
		if got := handleDurationRayNative48EA70([]byte{0x9e, kind, 0, 1, 0, 2, 0}, &slots, durationRayHooks48EA70{}); got != -1 {
			t.Fatalf("invalid kind %d consumed %d bytes", kind, got)
		}
	}
	packet := []byte{0x9e, 3, 0, 1, 0, 2, 0}
	if got := handleDurationRayNative48EA70(packet, &slots, durationRayHooks48EA70{
		connected: func() bool { return false },
	}); got != 7 {
		t.Fatalf("disconnected packet consumed %d bytes", got)
	}
	spawned := 0
	hooks := durationRayHooks48EA70{
		connected: func() bool { return true },
		byCode:    func(uint16) *client.Drawable { return nil },
		spawn:     func(int, image.Point) *client.Drawable { spawned++; return nil },
	}
	if got := handleDurationRayNative48EA70(packet, &slots, hooks); got != 7 || spawned != 0 {
		t.Fatalf("missing endpoints consumed %d bytes and spawned %d rays", got, spawned)
	}
	for i := range slots {
		slots[i].drawable = new(client.Drawable)
	}
	if got := handleDurationRayNative48EA70(packet, &slots, hooks); got != 7 || spawned != 0 {
		t.Fatalf("full list consumed %d bytes and spawned %d rays", got, spawned)
	}
}

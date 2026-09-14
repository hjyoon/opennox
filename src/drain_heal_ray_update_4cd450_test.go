package opennox

import (
	"encoding/binary"
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestDrainHealRayUpdateNativeWidth4CD450(t *testing.T) {
	from := &client.Drawable{PosVec: image.Pt(1000, 2000)}
	to := &client.Drawable{PosVec: image.Pt(1100, 2100)}
	ray := &client.Drawable{ZVal: 10}
	orb := new(client.Drawable)
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ray)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", ray)
	}
	setDurationRayDrawable48EA70(ray, 0, 0x1234, 0x9234)
	vp := &noxrender.Viewport{
		Screen: image.Rect(0, 0, 640, 480),
		World:  image.Rect(1000, 2000, 1640, 2480),
		Size:   image.Pt(640, 480),
	}
	random := []int{0, 2, -3, 7, 4}
	index := 0
	hooks := drainHealRayHooks4CD450{
		typeID: func(name string) int {
			if name != "HealOrb" {
				t.Fatalf("orb type %q", name)
			}
			return 71
		},
		random: func(_, _ int) int {
			v := random[index]
			index++
			return v
		},
		byCode: func(code uint16) *client.Drawable {
			switch code {
			case 0x1234:
				return from
			case 0x9234:
				return to
			default:
				t.Fatalf("ray net code %#x", code)
				return nil
			}
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ != 71 || pos != (image.Pt(1102, 2097)) {
				t.Fatalf("spawn = %d/%v", typ, pos)
			}
			return orb
		},
	}
	if got := updateDrainHealRay4CD450(vp, ray, "HealOrb", hooks); got != 1 {
		t.Fatalf("update returned %d", got)
	}
	if index != len(random) {
		t.Fatalf("used %d random values", index)
	}
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&orb.Union)), 13)
	if binary.LittleEndian.Uint16(payload[0:]) != 1000 || binary.LittleEndian.Uint16(payload[2:]) != 2000 ||
		payload[11] != 7 || payload[12] != 4 {
		t.Fatalf("orb payload = %x", payload)
	}
}

func TestDrainHealRayUpdateFixedCoordinatesAndClipping4CD450(t *testing.T) {
	ray := new(client.Drawable)
	orb := new(client.Drawable)
	ray.ZVal = 10
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&ray.Union)), 13)
	binary.LittleEndian.PutUint16(payload[5:], 300)
	binary.LittleEndian.PutUint16(payload[7:], 400)
	binary.LittleEndian.PutUint16(payload[9:], 500)
	binary.LittleEndian.PutUint16(payload[11:], 600)
	vp := &noxrender.Viewport{Screen: image.Rect(0, 0, 640, 480), Size: image.Pt(640, 480)}
	hooks := drainHealRayHooks4CD450{
		typeID: func(name string) int {
			if name != "DrainManaOrb" {
				t.Fatalf("orb type %q", name)
			}
			return 72
		},
		random: func(_, _ int) int { return 0 },
		byCode: func(uint16) *client.Drawable {
			t.Fatal("fixed-coordinate ray looked up net code")
			return nil
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ != 72 || pos != (image.Pt(500, 469)) {
				t.Fatalf("spawn = %d/%v", typ, pos)
			}
			return orb
		},
	}
	if got := updateDrainHealRay4CD450(vp, ray, "DrainManaOrb", hooks); got != 1 {
		t.Fatalf("update returned %d", got)
	}
	orbPayload := unsafe.Slice((*byte)(unsafe.Pointer(&orb.Union)), 13)
	if binary.LittleEndian.Uint16(orbPayload[0:]) != 300 || binary.LittleEndian.Uint16(orbPayload[2:]) != 400 {
		t.Fatalf("orb source = %x", orbPayload[:4])
	}
}

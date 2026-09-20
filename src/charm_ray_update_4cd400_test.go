package opennox

import (
	"encoding/binary"
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/client"
	"github.com/opennox/opennox/v1/client/noxrender"
)

func TestCharmRayUpdateNativeWidth4CD400(t *testing.T) {
	from := &client.Drawable{PosVec: image.Pt(1000, 2000)}
	to := &client.Drawable{PosVec: image.Pt(1100, 2100)}
	ray := &client.Drawable{ZVal: 10}
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(ray)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low address: %p", ray)
	}
	setDurationRayDrawable48EA70(ray, 0, 0x1234, 0x9234)
	vp := &noxrender.Viewport{
		Screen: image.Rect(0, 0, 640, 480),
		World:  image.Rect(1000, 2000, 1640, 2480),
		Size:   image.Pt(640, 480),
	}
	random := []int{0, 2, -3, 7, 4, 0, 5, 16, 8, 3}
	index := 0
	var orbs []*client.Drawable
	var positions []image.Point
	hooks := drainHealRayHooks4CD450{
		typeID: func(name string) int {
			if name != "CharmOrb" {
				t.Fatalf("orb type %q", name)
			}
			return 73
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
			if typ != 73 {
				t.Fatalf("spawn type = %d", typ)
			}
			orb := new(client.Drawable)
			orbs = append(orbs, orb)
			positions = append(positions, pos)
			return orb
		},
	}
	if got := updateCharmRay4CD400(vp, ray, hooks); got != 1 {
		t.Fatalf("update returned %d", got)
	}
	if index != len(random) {
		t.Fatalf("used %d random values", index)
	}
	if len(orbs) != 2 {
		t.Fatalf("spawned %d orbs", len(orbs))
	}
	if positions[0] != (image.Pt(1102, 2097)) || positions[1] != (image.Pt(1005, 2016)) {
		t.Fatalf("spawn positions = %v", positions)
	}
	assertCharmOrbPayload4CD400(t, orbs[0], image.Pt(1000, 2000), 7, 4)
	assertCharmOrbPayload4CD400(t, orbs[1], image.Pt(1100, 2100), 8, 3)
}

func TestCharmRayUpdateFixedCoordinatesAndClipping4CD400(t *testing.T) {
	ray := &client.Drawable{ZVal: 10}
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&ray.Union)), 13)
	binary.LittleEndian.PutUint16(payload[5:], 300)
	binary.LittleEndian.PutUint16(payload[7:], 400)
	binary.LittleEndian.PutUint16(payload[9:], 500)
	binary.LittleEndian.PutUint16(payload[11:], 600)
	vp := &noxrender.Viewport{Screen: image.Rect(0, 0, 640, 480), Size: image.Pt(640, 480)}
	random := []int{0, 0, 0, 6, 3, 50}
	index := 0
	orb := new(client.Drawable)
	hooks := drainHealRayHooks4CD450{
		typeID: func(name string) int {
			if name != "CharmOrb" {
				t.Fatalf("orb type %q", name)
			}
			return 74
		},
		random: func(_, _ int) int {
			v := random[index]
			index++
			return v
		},
		byCode: func(uint16) *client.Drawable {
			t.Fatal("fixed-coordinate ray looked up net code")
			return nil
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			if typ != 74 || pos != (image.Pt(500, 469)) {
				t.Fatalf("spawn = %d/%v", typ, pos)
			}
			return orb
		},
	}
	if got := updateCharmRay4CD400(vp, ray, hooks); got != 1 {
		t.Fatalf("update returned %d", got)
	}
	if index != len(random) {
		t.Fatalf("used %d random values", index)
	}
	assertCharmOrbPayload4CD400(t, orb, image.Pt(300, 400), 6, 3)
}

func assertCharmOrbPayload4CD400(t *testing.T, orb *client.Drawable, destination image.Point, radius, fade byte) {
	t.Helper()
	payload := unsafe.Slice((*byte)(unsafe.Pointer(&orb.Union)), 13)
	if got := image.Pt(int(binary.LittleEndian.Uint16(payload[0:])), int(binary.LittleEndian.Uint16(payload[2:]))); got != destination {
		t.Fatalf("orb destination = %v, want %v", got, destination)
	}
	if payload[11] != radius || payload[12] != fade {
		t.Fatalf("orb timing = %d/%d, want %d/%d", payload[11], payload[12], radius, fade)
	}
}

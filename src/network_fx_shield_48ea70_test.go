package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client"
)

func TestHandleShieldFXNative48EA70HighAddress(t *testing.T) {
	target := &client.Drawable{PosVec: image.Pt(100, 200)}
	shield := &client.Drawable{}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(target)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(shield)) <= uintptr(^uint32(0))) {
		t.Skipf("allocator returned a low drawable address: target=%p shield=%p", target, shield)
	}
	packet := []byte{byte(netmsg.MSG_FX_SHIELD), 0x34, 0x92, 2}
	spawned := 0
	hooks := shieldFXHooks48EA70{
		connected: func() bool { return true },
		byCode: func(code uint16) *client.Drawable {
			if code != 0x9234 {
				t.Fatalf("target code = %#x", code)
			}
			return target
		},
		typeID: func(index int, name string) int {
			if name != sphericalShieldNames48EA70[index] {
				t.Fatalf("type %d name = %q", index, name)
			}
			return 100 + index
		},
		eachNear: func(rect image.Rectangle, fnc func(*client.Drawable)) {
			if rect != image.Rect(90, 190, 111, 211) {
				t.Fatalf("scan rect = %v", rect)
			}
		},
		spawn: func(typ int, pos image.Point) *client.Drawable {
			spawned++
			if typ != 102 || pos != image.Pt(100, 203) {
				t.Fatalf("spawn = type %d at %v", typ, pos)
			}
			shield.TypeIDVal = uint32(typ)
			shield.PosVec = pos
			return shield
		},
	}
	if got := handleShieldFXNative48EA70(packet, hooks); got != 4 || spawned != 1 || shield.UnionEffect().Field_108 != 0x9234 {
		t.Fatalf("result=%d spawned=%d code=%#x", got, spawned, shield.UnionEffect().Field_108)
	}

	hooks.eachNear = func(_ image.Rectangle, fnc func(*client.Drawable)) { fnc(shield) }
	hooks.spawn = func(int, image.Point) *client.Drawable {
		spawned++
		t.Fatal("duplicate shield spawned")
		return nil
	}
	if got := handleShieldFXNative48EA70(packet, hooks); got != 4 || spawned != 1 {
		t.Fatalf("duplicate result=%d spawned=%d", got, spawned)
	}
}

func TestHandleShieldFXNative48EA70Guards(t *testing.T) {
	if got := handleShieldFXNative48EA70(make([]byte, 3), shieldFXHooks48EA70{}); got != -1 {
		t.Fatalf("short shield consumed %d", got)
	}
	packet := []byte{byte(netmsg.MSG_FX_SHIELD), 1, 0, 0}
	if got := handleShieldFXNative48EA70(packet, shieldFXHooks48EA70{connected: func() bool { return false }}); got != 4 {
		t.Fatalf("disconnected shield consumed %d", got)
	}
	spawned := 0
	hooks := shieldFXHooks48EA70{
		connected: func() bool { return true },
		byCode:    func(uint16) *client.Drawable { return nil },
		spawn: func(int, image.Point) *client.Drawable {
			spawned++
			return nil
		},
	}
	if got := handleShieldFXNative48EA70(packet, hooks); got != 4 || spawned != 0 {
		t.Fatalf("missing-target result=%d spawned=%d", got, spawned)
	}
	packet[3] = 9
	if got := handleShieldFXNative48EA70(packet, hooks); got != 4 || spawned != 0 {
		t.Fatalf("invalid-direction result=%d spawned=%d", got, spawned)
	}
}

func TestHandleShieldFXNative48EA70CenterDirectionSkips(t *testing.T) {
	target := &client.Drawable{PosVec: image.Pt(10, 20)}
	spawned := 0
	packet := []byte{byte(netmsg.MSG_FX_SHIELD), 1, 0, 4}
	got := handleShieldFXNative48EA70(packet, shieldFXHooks48EA70{
		connected: func() bool { return true },
		byCode:    func(uint16) *client.Drawable { return target },
		typeID:    func(index int, _ string) int { return index + 1 },
		eachNear:  func(image.Rectangle, func(*client.Drawable)) {},
		spawn: func(int, image.Point) *client.Drawable {
			spawned++
			return nil
		},
	})
	if got != 4 || spawned != 0 {
		t.Fatalf("center-direction result=%d spawned=%d", got, spawned)
	}
}

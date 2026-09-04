package opennox

import (
	"fmt"
	"image"
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client"
)

func TestPointSpriteFXSpecForOp48EA70(t *testing.T) {
	tests := []struct {
		op    netmsg.Op
		cache int
		name  string
		yoff  int
	}{
		{netmsg.MSG_FX_EXPLOSION, 0, "FireBoom", 0},
		{netmsg.MSG_FX_LESSER_EXPLOSION, 1, "MediumFireBoom", 0},
		{netmsg.MSG_FX_COUNTERSPELL_EXPLOSION, 2, "CounterspellBoom", 0},
		{netmsg.MSG_FX_THIN_EXPLOSION, 3, "ThinFireBoom", 0},
		{netmsg.MSG_FX_TELEPORT, 4, "TeleportPoof", 2},
		{netmsg.MSG_FX_DAMAGE_POOF, 5, "DamagePoof", 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := pointSpriteFXSpecForOp48EA70(tc.op)
			want := pointSpriteFXSpec48EA70{cache: tc.cache, name: tc.name, yoff: tc.yoff}
			if !ok || got != want {
				t.Fatalf("spec = %+v, ok=%t, want %+v, true", got, ok, want)
			}
		})
	}
	for _, op := range []netmsg.Op{netmsg.MSG_FX_SMOKE_BLAST, netmsg.MSG_FX_MANA_BOMB_CANCEL, 0xff} {
		if got, ok := pointSpriteFXSpecForOp48EA70(op); ok {
			t.Fatalf("unsupported opcode %#x returned spec %+v", byte(op), got)
		}
	}
}

func TestResolvePointSpriteFXTypeNative48EA70Cache(t *testing.T) {
	var cache int
	var calls []string
	lookup := func(name string) int {
		calls = append(calls, name)
		return 73
	}
	for i := 0; i < 2; i++ {
		if got := resolvePointSpriteFXTypeNative48EA70(&cache, "DamagePoof", lookup); got != 73 {
			t.Fatalf("resolved type = %d, want 73", got)
		}
	}
	if want := []string{"DamagePoof"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("lookups = %v, want %v", calls, want)
	}
}

func TestResolvePointSpriteFXTypeNative48EA70RetriesZero(t *testing.T) {
	var cache, calls int
	lookup := func(string) int {
		calls++
		return 0
	}
	resolvePointSpriteFXTypeNative48EA70(&cache, "DamagePoof", lookup)
	resolvePointSpriteFXTypeNative48EA70(&cache, "DamagePoof", lookup)
	if calls != 2 {
		t.Fatalf("lookups = %d, want 2 when the type remains unresolved", calls)
	}
}

func TestHandlePointSpriteFXNative48EA70Disconnected(t *testing.T) {
	var calls []string
	hooks := pointSpriteFXHooks48EA70{
		connected: func() bool {
			calls = append(calls, "connected")
			return false
		},
		typeID: func(pointSpriteFXSpec48EA70) int {
			calls = append(calls, "type")
			return 0
		},
	}
	if got := handlePointSpriteFXNative48EA70(netmsg.MSG_FX_DAMAGE_POOF, []byte{0x8b, 1, 0, 2, 0}, hooks); got != 5 {
		t.Fatalf("consumed bytes = %d, want 5", got)
	}
	if want := []string{"connected"}; !reflect.DeepEqual(calls, want) {
		t.Fatalf("callback order = %v, want %v", calls, want)
	}
}

func TestHandlePointSpriteFXNative48EA70HighAddress(t *testing.T) {
	tests := []struct {
		op   netmsg.Op
		name string
		pos  image.Point
	}{
		{netmsg.MSG_FX_EXPLOSION, "FireBoom", image.Pt(-32768, 32767)},
		{netmsg.MSG_FX_LESSER_EXPLOSION, "MediumFireBoom", image.Pt(-32768, 32767)},
		{netmsg.MSG_FX_COUNTERSPELL_EXPLOSION, "CounterspellBoom", image.Pt(-32768, 32767)},
		{netmsg.MSG_FX_THIN_EXPLOSION, "ThinFireBoom", image.Pt(-32768, 32767)},
		{netmsg.MSG_FX_TELEPORT, "TeleportPoof", image.Pt(-32768, 32769)},
		{netmsg.MSG_FX_DAMAGE_POOF, "DamagePoof", image.Pt(-32768, 32769)},
	}
	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dr := new(client.Drawable)
			if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(dr)) <= uintptr(^uint32(0)) {
				t.Skipf("allocator returned a low address: %p", dr)
			}
			var calls []string
			hooks := pointSpriteFXHooks48EA70{
				connected: func() bool {
					calls = append(calls, "connected")
					return true
				},
				typeID: func(spec pointSpriteFXSpec48EA70) int {
					calls = append(calls, fmt.Sprintf("type:%d:%s", spec.cache, spec.name))
					return 71 + i
				},
				spawn: func(typ int, pos image.Point) *client.Drawable {
					calls = append(calls, fmt.Sprintf("spawn:%d:%d:%d", typ, pos.X, pos.Y))
					return dr
				},
				activate: func(got *client.Drawable) {
					calls = append(calls, fmt.Sprintf("activate:%p", got))
					if got != dr {
						t.Fatalf("activated drawable = %p, want native pointer %p", got, dr)
					}
				},
			}
			data := []byte{byte(tc.op), 0x00, 0x80, 0xff, 0x7f, 0xaa}
			before := append([]byte(nil), data...)
			if got := handlePointSpriteFXNative48EA70(tc.op, data, hooks); got != 5 {
				t.Fatalf("consumed bytes = %d, want 5", got)
			}
			wantCalls := []string{
				"connected",
				fmt.Sprintf("type:%d:%s", i, tc.name),
				fmt.Sprintf("spawn:%d:%d:%d", 71+i, tc.pos.X, tc.pos.Y),
				fmt.Sprintf("activate:%p", dr),
			}
			if !reflect.DeepEqual(calls, wantCalls) {
				t.Fatalf("callback order = %v, want %v", calls, wantCalls)
			}
			if !reflect.DeepEqual(data, before) {
				t.Fatalf("packet mutated: got %x, want %x", data, before)
			}
		})
	}
}

func TestHandlePointSpriteFXNative48EA70NilAndInvalid(t *testing.T) {
	activateCalls := 0
	hooks := pointSpriteFXHooks48EA70{
		connected: func() bool { return true },
		typeID:    func(pointSpriteFXSpec48EA70) int { return 71 },
		spawn:     func(int, image.Point) *client.Drawable { return nil },
		activate:  func(*client.Drawable) { activateCalls++ },
	}
	if got := handlePointSpriteFXNative48EA70(netmsg.MSG_FX_DAMAGE_POOF, []byte{0x8b, 1, 0, 2, 0}, hooks); got != 5 {
		t.Fatalf("nil-spawn consumed bytes = %d, want 5", got)
	}
	if activateCalls != 0 {
		t.Fatalf("nil spawn activated %d drawables, want 0", activateCalls)
	}
	for n := 0; n < 5; n++ {
		if got := handlePointSpriteFXNative48EA70(netmsg.MSG_FX_DAMAGE_POOF, []byte{0x8b, 1, 0, 2, 0}[:n], pointSpriteFXHooks48EA70{}); got != -1 {
			t.Fatalf("%d-byte packet consumed bytes = %d, want -1", n, got)
		}
	}
	if got := handlePointSpriteFXNative48EA70(netmsg.MSG_FX_SMOKE_BLAST, []byte{0x8a, 1, 0, 2, 0}, pointSpriteFXHooks48EA70{}); got != -1 {
		t.Fatalf("unsupported opcode consumed bytes = %d, want -1", got)
	}
}

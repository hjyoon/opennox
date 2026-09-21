package opennox

import (
	"image"
	"testing"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	"github.com/opennox/opennox/v1/client"
)

func TestHandleSummonFXNative48EA70HighAddress(t *testing.T) {
	parent := &client.Drawable{}
	child := &client.Drawable{}
	if unsafe.Sizeof(uintptr(0)) == 8 && (uintptr(unsafe.Pointer(parent)) <= uintptr(^uint32(0)) || uintptr(unsafe.Pointer(child)) <= uintptr(^uint32(0))) {
		t.Skipf("allocator returned a low drawable address: parent=%p child=%p", parent, child)
	}
	packet := []byte{
		byte(netmsg.MSG_FX_SUMMON),
		0x34, 0x12, 0x78, 0x56,
		0xcd, 0xab,
		0x57, 0x13,
		0x21,
		0x68, 0x24,
	}
	got := handleSummonFXNative48EA70(packet, summonFXHooks48EA70{
		connected: func() bool { return true },
		typeID: func(name string) int {
			if name != "SummonEffect" {
				t.Fatalf("type name = %q", name)
			}
			return 77
		},
		spawnParent: func(typ int, pos image.Point) *client.Drawable {
			if typ != 77 || pos != image.Pt(0x1234, 0x5678) {
				t.Fatalf("parent spawn = %d at %v", typ, pos)
			}
			parent.TypeIDVal = uint32(typ)
			parent.PosVec = pos
			return parent
		},
		newChild: func(typ int) *client.Drawable {
			if typ != 0x1357 {
				t.Fatalf("child type = %#x", typ)
			}
			return child
		},
		direction: func(v byte) byte {
			if v != 0x21 {
				t.Fatalf("direction input = %#x", v)
			}
			return 6
		},
		frame: func() uint32 { return 9001 },
	})
	state := parent.UnionSummon()
	if got != 12 || state.Child != child || state.Lifetime != 0x2468 || state.ID != 0xabcd || parent.AnimStart != 9001 {
		t.Fatalf("summon result=%d state=%+v animStart=%d", got, state, parent.AnimStart)
	}
	if child.PosVec != image.Pt(0x1234, 0x5678) || child.AnimDir != 6 || child.AnimInd != 8 {
		t.Fatalf("child state = pos:%v dir:%d anim:%d", child.PosVec, child.AnimDir, child.AnimInd)
	}
}

func TestHandleSummonCancelFXNative48EA70HighAddress(t *testing.T) {
	decoy := &client.Drawable{TypeIDVal: 77}
	decoy.UnionSummon().ID = 0x1111
	parent := &client.Drawable{TypeIDVal: 77, PosVec: image.Pt(44, 55)}
	child := &client.Drawable{}
	parent.UnionSummon().ID = 0xabcd
	parent.UnionSummon().Child = child
	if unsafe.Sizeof(uintptr(0)) == 8 && uintptr(unsafe.Pointer(child)) <= uintptr(^uint32(0)) {
		t.Skipf("allocator returned a low child address: %p", child)
	}
	var spark image.Point
	var deletedChild, deletedParent *client.Drawable
	got := handleSummonCancelFXNative48EA70([]byte{byte(netmsg.MSG_FX_SUMMON_CANCEL), 0xcd, 0xab}, summonFXHooks48EA70{
		connected: func() bool { return true },
		typeID: func(name string) int {
			if name != "SummonEffect" {
				t.Fatalf("type name = %q", name)
			}
			return 77
		},
		each: func(fnc func(*client.Drawable)) {
			fnc(decoy)
			fnc(parent)
		},
		pointSpark:   func(pos image.Point) { spark = pos },
		deleteChild:  func(dr *client.Drawable) { deletedChild = dr },
		deleteParent: func(dr *client.Drawable) { deletedParent = dr },
	})
	if got != 3 || spark != image.Pt(44, 55) || deletedChild != child || deletedParent != parent || parent.UnionSummon().Child != nil {
		t.Fatalf("cancel result=%d spark=%v child=%p parent=%p remaining=%p", got, spark, deletedChild, deletedParent, parent.UnionSummon().Child)
	}
}

func TestHandleSummonFXNative48EA70Guards(t *testing.T) {
	if got := handleSummonFXNative48EA70(make([]byte, 11), summonFXHooks48EA70{}); got != -1 {
		t.Fatalf("short summon consumed %d", got)
	}
	if got := handleSummonCancelFXNative48EA70(make([]byte, 2), summonFXHooks48EA70{}); got != -1 {
		t.Fatalf("short cancel consumed %d", got)
	}
	if got := handleSummonFXNative48EA70(make([]byte, 12), summonFXHooks48EA70{connected: func() bool { return false }}); got != 12 {
		t.Fatalf("disconnected summon consumed %d", got)
	}
	if got := handleSummonCancelFXNative48EA70(make([]byte, 3), summonFXHooks48EA70{connected: func() bool { return false }}); got != 3 {
		t.Fatalf("disconnected cancel consumed %d", got)
	}
}

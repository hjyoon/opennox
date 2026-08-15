package opennox

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func TestUnitRemoveChildNativeLayouts4EC470(t *testing.T) {
	type layout struct {
		objectSize uintptr
		owner      uintptr
		nextOwned  uintptr
		firstOwned uintptr
	}
	var want layout
	switch unsafe.Sizeof(uintptr(0)) {
	case 4:
		want = layout{780, 508, 512, 516}
	case 8:
		want = layout{928, 552, 560, 568}
	default:
		t.Fatalf("unsupported pointer width %d", unsafe.Sizeof(uintptr(0)))
	}
	got := layout{
		objectSize: unsafe.Sizeof(server.Object{}),
		owner:      unsafe.Offsetof(server.Object{}.ObjOwner),
		nextOwned:  unsafe.Offsetof(server.Object{}.Field128),
		firstOwned: unsafe.Offsetof(server.Object{}.Field129),
	}
	if got != want {
		t.Fatalf("native layout = %+v, want %+v", got, want)
	}
}

func TestUnitRemoveChildNative4EC470ClearsWholeOwnedList(t *testing.T) {
	foreignOwner := &server.Object{}
	first := &server.Object{ObjOwner: foreignOwner}
	second := &server.Object{}
	parent := &server.Object{Field129: first}
	first.Field128 = second

	unitRemoveChild4EC470(parent)
	if parent.Field129 != nil || first.ObjOwner != nil || first.Field128 != nil || second.ObjOwner != nil || second.Field128 != nil {
		t.Fatalf("links were not cleared: parent=%p first=(%p,%p) second=(%p,%p)",
			parent.Field129, first.ObjOwner, first.Field128, second.ObjOwner, second.Field128)
	}
	unitRemoveChild4EC470(parent)
	unitRemoveChild4EC470(nil)
}

func TestUnitRemoveChildNative4EC470SelfCycle(t *testing.T) {
	parent := &server.Object{}
	child := &server.Object{ObjOwner: parent}
	parent.Field129 = child
	child.Field128 = child

	unitRemoveChild4EC470(parent)
	if parent.Field129 != nil || child.ObjOwner != nil || child.Field128 != nil {
		t.Fatalf("self-cycle links = parent %p owner %p next %p", parent.Field129, child.ObjOwner, child.Field128)
	}
}

package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestScriptObjectLookupNative4DA4F0(t *testing.T) {
	s := new(server.Server)
	old := GetServer
	GetServer = func() Server { return &chatLegacyServer528AC0{srv: s} }
	t.Cleanup(func() { GetServer = old })
	newObject := func(name string) *server.Object {
		obj, free := alloc.NewHigh(server.Object{})
		t.Cleanup(free)
		id, freeID := alloc.CStringHigh(name)
		t.Cleanup(freeID)
		obj.IDPtr = unsafe.Pointer(id)
		return obj
	}
	owner, freeOwner := alloc.NewHigh(server.Object{})
	t.Cleanup(freeOwner)
	key := newObject("Hall:Key")
	nested := newObject("Hidden")
	second := newObject("Map:Key")
	deepName := newObject("Outer:Inner:Door")
	emptyName := newObject("")
	pendingDuplicate := newObject("Hall:Key")
	pending := newObject("Future:Later")

	owner.InvFirstItem = key
	key.InvFirstItem = nested
	owner.ObjNext, second.ObjNext, deepName.ObjNext = second, deepName, emptyName
	pendingDuplicate.ObjNext = pending
	s.Objs.List, s.Objs.Pending = owner, pendingDuplicate

	for _, tc := range []struct {
		query string
		want  *server.Object
	}{
		{"Key", key},
		{"Hall:Key", key},
		{"Map:Key", second},
		{"key", nil},
		{"Hidden", nil},
		{"Outer:Inner:Door", deepName},
		{"Inner:Door", nil},
		{"Door", nil},
		{"Later", pending},
		{"", emptyName},
		{"Key\x00hidden", key},
	} {
		for _, call := range []struct {
			name   string
			lookup func(string) *server.Object
		}{
			{"Go", Nox_xxx_getObjectByScrName_4DA4F0},
			{"C", callScriptObjectLookupC},
		} {
			if got := call.lookup(tc.query); got != tc.want {
				t.Errorf("%s lookup %q = %p, want %p", call.name, tc.query, got, tc.want)
			}
		}
	}
}

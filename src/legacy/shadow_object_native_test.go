package legacy

import (
	"container/list"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func TestShadowObjectNativeCEntries(t *testing.T) {
	old := shadowObjects
	shadowObjects = list.New()
	t.Cleanup(func() { shadowObjects = old })
	var objs [4]*server.Object
	for i := range objs {
		obj, free := alloc.NewHigh(server.Object{})
		t.Cleanup(free)
		objs[i] = obj
		obj.ObjFlags = object.FlagActive
		obj.Field117, obj.Field118 = 17, 18
	}
	first, middle, last, skipped := objs[0], objs[1], objs[2], objs[3]
	check := func(obj *server.Object, flags object.Flags, a, b uint32) {
		t.Helper()
		if obj.ObjFlags != flags || obj.Field117 != a || obj.Field118 != b {
			t.Fatalf("object %p holds flags %#x and fields %d/%d, want %#x and %d/%d",
				obj, obj.ObjFlags, obj.Field117, obj.Field118, flags, a, b)
		}
	}
	order := func(want ...*server.Object) {
		t.Helper()
		if shadowObjects.Len() != len(want) {
			t.Fatalf("shadow list length %d, want %d", shadowObjects.Len(), len(want))
		}
		link := shadowObjects.Front()
		for i, obj := range want {
			if got := link.Value; got != obj {
				t.Fatalf("shadow list member %d = %p, want %p", i, got, obj)
			}
			link = link.Next()
		}
	}
	skipped.ObjFlags |= object.FlagTransient
	if got := addObjectShadowC(skipped); got != unsafe.Pointer(skipped) {
		t.Fatalf("transient add returned %p, want %p", got, skipped)
	}
	check(skipped, object.FlagActive|object.FlagTransient, 17, 18)
	order()
	for _, obj := range objs[:3] {
		if got := addObjectShadowC(obj); got != unsafe.Pointer(obj) {
			t.Fatalf("add returned %p, want %p", got, obj)
		}
		check(obj, object.FlagActive|object.FlagShadow, 0, 0)
	}
	order(last, middle, first)
	middle.Field117, middle.Field118 = 23, 24
	Nox_xxx_unitNewAddShadow_4DA9A0(middle)
	check(middle, object.FlagActive|object.FlagShadow, 23, 24)
	order(last, middle, first)
	for _, obj := range []*server.Object{middle, first, last} {
		if got := removeObjectShadowC(obj); got != unsafe.Pointer(obj) {
			t.Fatalf("remove returned %p, want %p", got, obj)
		}
		check(obj, object.FlagActive, 0, 0)
		switch obj {
		case middle:
			order(last, first)
		case first:
			order(last)
		case last:
			order()
		}
	}
	skipped.ObjFlags |= object.FlagShadow
	Nox_xxx_action_4DA9F0(skipped)
	check(skipped, object.FlagActive|object.FlagTransient, 0, 0)
	skipped.Field117, skipped.Field118 = 31, 32
	removeObjectShadowC(skipped)
	check(skipped, object.FlagActive|object.FlagTransient, 31, 32)
	order()
}

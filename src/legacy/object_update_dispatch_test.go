package legacy

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestMonsterUpdateDispatchStaysInGo(t *testing.T) {
	callback, size, ok := server.ObjectUpdateHandler("MonsterUpdate")
	if !ok || callback == nil || size != unsafe.Sizeof(server.MonsterUpdateData{}) {
		t.Fatalf("MonsterUpdate registration = %p/%d/%t", callback, size, ok)
	}

	original := Nox_xxx_unitUpdateMonster_50A5C0
	t.Cleanup(func() {
		Nox_xxx_unitUpdateMonster_50A5C0 = original
	})

	var called *server.Object
	Nox_xxx_unitUpdateMonster_50A5C0 = func(obj *server.Object) {
		called = obj
	}

	// Keep a live Go pointer in the object. A regression to the legacy C
	// trampoline would make cgo reject this pointer graph before the callback.
	owner := &server.Object{}
	obj := &server.Object{
		ObjClass: object.ClassMonster,
		ObjOwner: owner,
	}
	server.CallObjectUpdate(callback, obj)
	if called != obj {
		t.Fatalf("MonsterUpdate called with %p, want %p", called, obj)
	}
}

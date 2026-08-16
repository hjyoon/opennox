package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/types"
)

func TestObjectCallDropUsesDroppedItemHandler(t *testing.T) {
	var token byte
	ptr := unsafe.Pointer(&token)
	actor := &Object{}
	item := &Object{Drop: DropFuncPtr{Ptr: ptr}}
	pos := types.Pointf{X: 12.5, Y: -3.25}
	called := false
	objDrop.Register(ptr, func(gotActor, gotItem *Object, gotPos *types.Pointf) int32 {
		called = true
		if gotActor != actor || gotItem != item || gotPos == nil || *gotPos != pos {
			t.Fatalf("drop callback = (%p, %p, %p), want (%p, %p, %#v)", gotActor, gotItem, gotPos, actor, item, pos)
		}
		return -1
	})
	if !actor.CallDrop(item, pos) || !called {
		t.Fatal("dropped item's callback was not invoked")
	}
	if actor.CallDrop(nil, pos) {
		t.Fatal("nil item drop succeeded")
	}
}

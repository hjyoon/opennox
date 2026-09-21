package server

import (
	"bytes"
	"fmt"
	"testing"
	"unsafe"
)

func TestObjectPickupHandlerReturnsExactRegistration(t *testing.T) {
	const name = "ObjectPickupHandlerTest"
	if _, ok := pickupFuncs[name]; ok {
		t.Fatalf("test pickup handler %q is already registered", name)
	}
	var storage byte
	want := unsafe.Pointer(&storage)
	pickupFuncs[name] = want
	t.Cleanup(func() { delete(pickupFuncs, name) })

	got, ok := ObjectPickupHandler(name)
	if !ok || got.Ptr != want {
		t.Fatalf("ObjectPickupHandler(%q) = %p/%t, want %p/true", name, got.Ptr, ok, want)
	}
	if got, ok := ObjectPickupHandler(name + "Missing"); ok || got.Ptr != nil {
		t.Fatalf("missing ObjectPickupHandler = %p/%t, want nil/false", got.Ptr, ok)
	}
}

func TestObjectUpdateHandlerReturnsExactRegistration(t *testing.T) {
	const name = "ObjectUpdateHandlerTest"
	if _, ok := updateFuncs[name]; ok {
		t.Fatalf("test update handler %q is already registered", name)
	}
	var storage byte
	wantPtr := unsafe.Pointer(&storage)
	wantSize := uintptr(4)
	updateFuncs[name] = objectDefFunc{Func: wantPtr, DataSize: wantSize}
	t.Cleanup(func() { delete(updateFuncs, name) })

	gotPtr, gotSize, ok := ObjectUpdateHandler(name)
	if !ok || gotPtr != wantPtr || gotSize != wantSize {
		t.Fatalf("ObjectUpdateHandler(%q) = %p/%d/%t, want %p/%d/true", name, gotPtr, gotSize, ok, wantPtr, wantSize)
	}
	if gotPtr, gotSize, ok := ObjectUpdateHandler(name + "Missing"); ok || gotPtr != nil || gotSize != 0 {
		t.Fatalf("missing ObjectUpdateHandler = %p/%d/%t, want nil/0/false", gotPtr, gotSize, ok)
	}
}

func TestObjectInitDispatchPreservesNativePointer(t *testing.T) {
	const name = "ObjectInitDispatchTest"
	if _, ok := initFuncs[name]; ok {
		t.Fatalf("test init handler %q is already registered", name)
	}
	callback := unsafe.Pointer(new(byte))
	wantSize := uintptr(12)
	var got *Object
	RegisterObjectInitGo(name, callback, func(obj *Object) {
		got = obj
	}, wantSize)
	t.Cleanup(func() { delete(initFuncs, name) })

	gotCallback, gotSize, ok := ObjectInitHandler(name)
	if !ok || gotCallback != callback || gotSize != wantSize {
		t.Fatalf("ObjectInitHandler(%q) = %p/%d/%t, want %p/%d/true",
			name, gotCallback, gotSize, ok, callback, wantSize)
	}

	owner := new(Object)
	obj := &Object{ObjOwner: owner, Init: callback}
	obj.CallInit()
	if got != obj || got.ObjOwner != owner {
		t.Fatalf("init object = %p owner %p, want %p owner %p", got, got.ObjOwner, obj, owner)
	}

	CallObjectInit(nil, obj)
	if gotCallback, gotSize, ok := ObjectInitHandler(name + "Missing"); ok || gotCallback != nil || gotSize != 0 {
		t.Fatalf("missing ObjectInitHandler = %p/%d/%t, want nil/0/false", gotCallback, gotSize, ok)
	}
}

func TestObjectCollideDispatchPreservesNativePointers(t *testing.T) {
	const name = "ObjectCollideDispatchTest"
	if _, ok := collideFuncs[name]; ok {
		t.Fatalf("test collide handler %q is already registered", name)
	}
	callback := unsafe.Pointer(new(byte))
	wantSize := uintptr(24)
	var gotObject, gotOther *Object
	var gotCollision unsafe.Pointer
	RegisterObjectCollideGo(name, callback, func(obj, other *Object, collision unsafe.Pointer) {
		gotObject = obj
		gotOther = other
		gotCollision = collision
	}, wantSize)
	t.Cleanup(func() { delete(collideFuncs, name) })

	gotCallback, gotSize, ok := ObjectCollideHandler(name)
	if !ok || gotCallback != callback || gotSize != wantSize {
		t.Fatalf("ObjectCollideHandler(%q) = %p/%d/%t, want %p/%d/true",
			name, gotCallback, gotSize, ok, callback, wantSize)
	}

	owner := new(Object)
	obj := &Object{ObjOwner: owner, Collide: callback}
	other := new(Object)
	collision := new(byte)
	CallObjectCollide(obj.Collide, obj, other, unsafe.Pointer(collision))
	if gotObject != obj || gotOther != other || gotCollision != unsafe.Pointer(collision) {
		t.Fatalf("collide args = (%p, %p, %p), want (%p, %p, %p)",
			gotObject, gotOther, gotCollision, obj, other, collision)
	}
	if gotObject.ObjOwner != owner {
		t.Fatalf("object owner = %p, want %p", gotObject.ObjOwner, owner)
	}

	CallObjectCollide(nil, obj, other, unsafe.Pointer(collision))
}

func TestCObjectUpdateTraceIdentifiesRegisteredHandler(t *testing.T) {
	const name = "ObjectUpdateTraceTest"
	var storage byte
	ptr := unsafe.Pointer(&storage)
	updateFuncs[name] = objectDefFunc{Func: ptr}
	t.Cleanup(func() { delete(updateFuncs, name) })
	if got := objectUpdateName(ptr); got != name {
		t.Fatalf("objectUpdateName(%p) = %q, want %q", ptr, got, name)
	}
	var updateData byte
	obj := &Object{TypeInd: 7, Extent: 42, UpdateData: unsafe.Pointer(&updateData)}
	var buf bytes.Buffer
	writeCObjectUpdateTrace(&buf, name, ptr, obj)
	want := fmt.Sprintf("NOX_C_UPDATE name=%q callback=%p object=%p data=%p type=7 extent=42\n",
		name, ptr, obj, obj.UpdateData)
	if got := buf.String(); got != want {
		t.Fatalf("C update trace = %q, want %q", got, want)
	}
}

func TestObjectDeathHandlerReturnsExactRegistration(t *testing.T) {
	const name = "ObjectDeathHandlerTest"
	if _, ok := deathFuncs[name]; ok {
		t.Fatalf("test death handler %q is already registered", name)
	}
	var storage byte
	wantPtr := unsafe.Pointer(&storage)
	wantSize := uintptr(132)
	deathFuncs[name] = objectDefFunc{Func: wantPtr, DataSize: wantSize}
	t.Cleanup(func() { delete(deathFuncs, name) })

	gotPtr, gotSize, ok := ObjectDeathHandler(name)
	if !ok || gotPtr != wantPtr || gotSize != wantSize {
		t.Fatalf("ObjectDeathHandler(%q) = %p/%d/%t, want %p/%d/true", name, gotPtr, gotSize, ok, wantPtr, wantSize)
	}
	if gotPtr, gotSize, ok := ObjectDeathHandler(name + "Missing"); ok || gotPtr != nil || gotSize != 0 {
		t.Fatalf("missing ObjectDeathHandler = %p/%d/%t, want nil/0/false", gotPtr, gotSize, ok)
	}
}

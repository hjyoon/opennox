package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func defaultOwnCollideNativeDeps4EA2C0() ownCollideNativeDeps4EA2C0 {
	return ownCollideNativeDeps4EA2C0{
		frame: func() uint32 { return 0 },
		setOwner: func(owner, obj *Object) {
			obj.ObjOwner = owner
		},
	}
}

func TestOwnCollide4EA2C0NativeLayout(t *testing.T) {
	wantClass := uintptr(8)
	wantFrame := uintptr(136)
	wantOwner := uintptr(508)
	wantSize := uintptr(780)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantFrame = 140
		wantOwner = 552
		wantSize = 928
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object size", unsafe.Sizeof(Object{}), wantSize},
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.Field34", unsafe.Offsetof(Object{}.Field34), wantFrame},
		{"Object.ObjOwner", unsafe.Offsetof(Object{}.ObjOwner), wantOwner},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestOwnCollideNative4EA2C0UsesNativeFieldsAndOrder(t *testing.T) {
	source := &Object{Field34: 0x11111111}
	target := &Object{ObjClass: object.ClassPlayer}
	events := make([]string, 0, 2)
	deps := defaultOwnCollideNativeDeps4EA2C0()
	deps.frame = func() uint32 {
		events = append(events, "frame")
		return 0xfedcba98
	}
	deps.setOwner = func(owner, obj *Object) {
		events = append(events, "set-owner")
		if owner != target || obj != source || obj.Field34 != 0xfedcba98 {
			t.Fatalf("set owner = %p/%p frame %#x", owner, obj, obj.Field34)
		}
		obj.ObjOwner = owner
	}
	ownCollideNative4EA2C0(source, target, deps)
	if !reflect.DeepEqual(events, []string{"frame", "set-owner"}) {
		t.Fatalf("events = %v", events)
	}
	if source.ObjOwner != target || source.Field34 != 0xfedcba98 {
		t.Fatalf("source = owner %p frame %#x", source.ObjOwner, source.Field34)
	}
}

func TestOwnCollide4EA2C0ServerBinding(t *testing.T) {
	s := &Server{}
	s.SetFrame(0x89abcdef)
	source := &Object{Field34: 0x11223344}
	target := &Object{ObjClass: object.ClassPlayer}
	s.OwnCollide4EA2C0(source, target)
	if source.Field34 != 0x89abcdef || source.ObjOwner != target {
		t.Fatalf("source = frame %#x owner %p", source.Field34, source.ObjOwner)
	}
	if target.Field129 != source || source.Field128 != nil {
		t.Fatalf("owner list = head %p next %p", target.Field129, source.Field128)
	}
}

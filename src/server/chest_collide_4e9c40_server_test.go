package server

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestChestCollide4E9C40NativeLayout(t *testing.T) {
	wantClass := uintptr(8)
	wantSubclass := uintptr(12)
	wantFlags := uintptr(16)
	wantNextItem := uintptr(496)
	wantFirstItem := uintptr(504)
	wantDeath := uintptr(724)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantSubclass = 16
		wantFlags = 20
		wantNextItem = 528
		wantFirstItem = 544
		wantDeath = 824
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjSubClass", unsafe.Offsetof(Object{}.ObjSubClass), wantSubclass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
		{"Object.InvNextItem", unsafe.Offsetof(Object{}.InvNextItem), wantNextItem},
		{"Object.InvFirstItem", unsafe.Offsetof(Object{}.InvFirstItem), wantFirstItem},
		{"Object.Death", unsafe.Offsetof(Object{}.Death), wantDeath},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestChestCollideNative4E9C40UsesPointersAndLiveDeath(t *testing.T) {
	first := &Object{ObjClass: object.ClassKey}
	key := &Object{ObjClass: object.ClassKey}
	first.InvNextItem = key
	target := &Object{ObjClass: object.ClassPlayer, InvFirstItem: first}
	oldDeath := unsafe.Pointer(new(byte))
	liveDeath := unsafe.Pointer(new(byte))
	source := &Object{ObjSubClass: 0x100, Death: oldDeath}
	collision := [2]float32{3.5, -8.25}
	var events []string

	chestCollideNative4E9C40(source, target, unsafe.Pointer(&collision[0]), chestCollideNativeDeps4E9C40{
		gameFlagsCheck: func(uint32) int32 { events = append(events, "quest"); return 1 },
		loadTypeName: func(item *Object) string {
			events = append(events, "name")
			if item == first {
				return "GoldKey"
			}
			return "SilverKey"
		},
		delayedDelete: func(item *Object) {
			events = append(events, "delete")
			if item != key {
				t.Fatalf("deleted item = %p, want %p", item, key)
			}
			source.Death = liveDeath
		},
		audio: func(id uint32, obj *Object) {
			events = append(events, "audio")
			if id != chestCollideUnlockAudio4E9C40 || obj != source {
				t.Fatalf("audio = %d/%p", id, obj)
			}
		},
		callDeath: func(death unsafe.Pointer, obj *Object) {
			events = append(events, "death")
			if death != liveDeath || obj != source {
				t.Fatalf("death = %p/%p, want %p/%p", death, obj, liveDeath, source)
			}
		},
		chestOpen: func(gotSource, gotTarget *Object) {
			events = append(events, "open")
			if gotSource != source || gotTarget != target {
				t.Fatalf("open = %p/%p", gotSource, gotTarget)
			}
		},
		dropAllItems: func(got *Object) {
			events = append(events, "drop")
			if got != source {
				t.Fatalf("drop source = %p", got)
			}
		},
	})

	want := []string{"quest", "name", "name", "delete", "audio", "death", "open", "drop"}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if collision != [2]float32{3.5, -8.25} {
		t.Fatalf("collision changed to %#v", collision)
	}
}

func TestChestCollideNative4E9C40NilSourceFaultsOnlyAfterPlayerGate(t *testing.T) {
	nonPlayer := &Object{}
	chestCollideNative4E9C40(nil, nonPlayer, nil, chestCollideNativeDeps4E9C40{})

	defer func() {
		if recover() == nil {
			t.Fatal("nil source did not fault after valid Player target")
		}
	}()
	chestCollideNative4E9C40(nil, &Object{ObjClass: object.ClassPlayer}, nil, chestCollideNativeDeps4E9C40{})
}

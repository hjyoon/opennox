package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestDropEligibilityNative4EDCD0ObjectLayout(t *testing.T) {
	wantClass := uintptr(8)
	wantFlags := uintptr(16)
	if unsafe.Sizeof(uintptr(0)) == 8 {
		wantClass = 12
		wantFlags = 20
	}
	checks := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"Object.ObjClass", unsafe.Offsetof(Object{}.ObjClass), wantClass},
		{"Object.ObjFlags", unsafe.Offsetof(Object{}.ObjFlags), wantFlags},
	}
	for _, check := range checks {
		if check.got != check.want {
			t.Errorf("%s offset = %d, want %d", check.name, check.got, check.want)
		}
	}
}

func TestDropEligibilityNative4EDCD0Branches(t *testing.T) {
	unit := object.ClassPlayer | object.ClassMonster
	tests := []struct {
		name  string
		owner *Object
		item  *Object
		want  int32
	}{
		{"destroyed unit no-drop", &Object{ObjClass: unit}, &Object{ObjFlags: object.FlagDestroyed | object.Flags(dropEligibilityNoDrop4EDCD0)}, 1},
		{"non-unit no-drop", &Object{ObjClass: object.ClassImmobile}, &Object{ObjFlags: object.Flags(dropEligibilityNoDrop4EDCD0)}, 1},
		{"player ordinary", &Object{ObjClass: object.ClassPlayer}, new(Object), 1},
		{"monster ordinary", &Object{ObjClass: object.ClassMonster}, new(Object), 1},
		{"player no-drop", &Object{ObjClass: object.ClassPlayer}, &Object{ObjFlags: object.Flags(dropEligibilityNoDrop4EDCD0)}, 0},
		{"monster no-drop", &Object{ObjClass: object.ClassMonster}, &Object{ObjFlags: object.Flags(dropEligibilityNoDrop4EDCD0)}, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := dropEligibilityNative4EDCD0(test.owner, test.item); got != test.want {
				t.Fatalf("result = %d, want %d", got, test.want)
			}
		})
	}
}

func TestDropEligibilityNative4EDCD0FaultOrder(t *testing.T) {
	destroyed := &Object{ObjFlags: object.FlagDestroyed}
	if got := dropEligibilityNative4EDCD0(nil, destroyed); got != 1 {
		t.Fatalf("destroyed item with nil owner = %d, want 1", got)
	}

	assertPanics := func(t *testing.T, call func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatal("call did not panic")
			}
		}()
		call()
	}
	t.Run("nil item precedes owner", func(t *testing.T) {
		assertPanics(t, func() { _ = dropEligibilityNative4EDCD0(nil, nil) })
	})
	t.Run("live item reaches nil owner", func(t *testing.T) {
		assertPanics(t, func() { _ = dropEligibilityNative4EDCD0(nil, new(Object)) })
	})
}

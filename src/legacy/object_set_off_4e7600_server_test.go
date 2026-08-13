package legacy

import (
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestObjectSetOffNative4E7600UpdatesNativeObject(t *testing.T) {
	obj := &server.Object{
		ObjClass: object.ClassClientPersist | object.ClassElevator | object.ClassFire,
		ObjFlags: object.FlagEnabled | object.FlagActive,
		Field37:  0x13579bdf,
	}
	audioCalls := 0
	got := objectSetOffNative4E7600(obj, func(got *server.Object) {
		audioCalls++
		if got != obj {
			t.Fatal("audio received a different object")
		}
	})
	if audioCalls != 1 {
		t.Fatalf("audio calls = %d, want 1", audioCalls)
	}
	if got != uint32(obj.ObjFlags) {
		t.Fatalf("result = %#x, want stored flags %#x", got, obj.ObjFlags)
	}
	if obj.ObjFlags.Has(object.FlagEnabled) || !obj.ObjFlags.Has(object.FlagNoCollide) || !obj.ObjFlags.Has(object.FlagActive) {
		t.Fatalf("flags = %#x, want Active+NoCollide and no Enabled", obj.ObjFlags)
	}
	if obj.Field37 != 0x13579bdf {
		t.Fatalf("neighboring field changed: %#x", obj.Field37)
	}
}

func TestObjectSetOffNative4E7600NilFaultsBeforeAudio(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil object returned without a panic")
		}
	}()
	objectSetOffNative4E7600(nil, func(*server.Object) {
		t.Fatal("nil object reached audio")
	})
}

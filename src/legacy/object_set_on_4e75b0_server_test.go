package legacy

import (
	"testing"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func TestObjectSetOnNative4E75B0UpdatesNativeObject(t *testing.T) {
	obj := &server.Object{
		ObjClass: object.ClassClientPersist | object.ClassElevator | object.ClassFire,
		ObjFlags: object.FlagActive | object.FlagNoCollide,
		Field37:  0x13579bdf,
	}
	audioCalls := 0
	helperCalls := 0
	got := objectSetOnNative4E75B0(
		obj,
		func(got *server.Object) {
			audioCalls++
			if got != obj {
				t.Fatal("audio received a different object")
			}
		},
		func(got *server.Object) byte {
			helperCalls++
			if got != obj {
				t.Fatal("collide/update helper received a different object")
			}
			return 0xfe
		},
	)
	if got != 0xfe || audioCalls != 1 || helperCalls != 1 {
		t.Fatalf("result/audio/helper = (%#x, %d, %d), want (0xfe, 1, 1)", got, audioCalls, helperCalls)
	}
	if !obj.ObjFlags.Has(object.FlagEnabled) || obj.ObjFlags.Has(object.FlagNoCollide) || !obj.ObjFlags.Has(object.FlagActive) {
		t.Fatalf("flags = %#x, want Enabled+Active and no NoCollide", obj.ObjFlags)
	}
	if obj.Field37 != 0x13579bdf {
		t.Fatalf("neighboring field changed: %#x", obj.Field37)
	}
}

func TestObjectSetOnNative4E75B0MissileReturnsClassByte(t *testing.T) {
	obj := &server.Object{
		ObjClass: object.ClassClientPersist | object.ClassMissile,
		ObjFlags: object.FlagEnabled | object.FlagNoCollide,
	}
	got := objectSetOnNative4E75B0(
		obj,
		func(*server.Object) { t.Fatal("enabled Missile played elevator audio") },
		func(*server.Object) byte {
			t.Fatal("Missile called collide/update helper")
			return 0
		},
	)
	if got != byte(object.ClassMissile) {
		t.Fatalf("result = %#x, want Missile class byte %#x", got, byte(object.ClassMissile))
	}
	if !obj.ObjFlags.Has(object.FlagNoCollide) {
		t.Fatal("class outside 0x10042000 unexpectedly cleared NoCollide")
	}
}

func TestObjectSetOnNative4E75B0NilFaultsBeforeCallbacks(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil object returned without a panic")
		}
	}()
	objectSetOnNative4E75B0(
		nil,
		func(*server.Object) { t.Fatal("nil object reached audio") },
		func(*server.Object) byte {
			t.Fatal("nil object reached collide/update helper")
			return 0
		},
	)
}

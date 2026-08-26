package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestSwitchUpdate53B320Enabled(t *testing.T) {
	t.Run("collidable", func(t *testing.T) {
		unit := &Object{ObjFlags: object.FlagActive | object.FlagEnabled | object.FlagNoCollide}
		unit.Collide = unsafe.Pointer(uintptr(1))
		var queued *Object
		got := new(Server).SwitchUpdate53B320(unit, SwitchUpdateRuntime53B320{
			QueueCollision: func(obj *Object) { queued = obj },
		})
		if unit.ObjFlags.Has(object.FlagNoCollide) || queued != unit {
			t.Fatalf("flags/queued = %#x/%p, want NoCollide clear and %p", unit.ObjFlags, queued, unit)
		}
		if got != uint8(unit.ObjFlags) {
			t.Fatalf("return = %#x, want %#x", got, uint8(unit.ObjFlags))
		}
	})
	t.Run("no collide callback", func(t *testing.T) {
		unit := &Object{ObjFlags: object.FlagEnabled | object.FlagNoCollide}
		calls := 0
		new(Server).SwitchUpdate53B320(unit, SwitchUpdateRuntime53B320{
			QueueCollision: func(*Object) { calls++ },
		})
		if unit.ObjFlags.Has(object.FlagNoCollide) || calls != 0 {
			t.Fatalf("flags/calls = %#x/%d, want NoCollide clear and 0", unit.ObjFlags, calls)
		}
	})
}

func TestSwitchUpdate53B320Disabled(t *testing.T) {
	t.Run("first update synchronizes", func(t *testing.T) {
		unit := &Object{ObjFlags: object.FlagActive}
		got := new(Server).SwitchUpdate53B320(unit, SwitchUpdateRuntime53B320{})
		if unit.Field33 != 1 || unit.Field38 != ^uint32(0) || !unit.ObjFlags.Has(object.FlagNoCollide) {
			t.Fatalf("state = field33:%d field38:%#x flags:%#x", unit.Field33, unit.Field38, unit.ObjFlags)
		}
		if got != uint8(unit.ObjFlags) {
			t.Fatalf("return = %#x, want %#x", got, uint8(unit.ObjFlags))
		}
	})
	t.Run("later update keeps sync marker", func(t *testing.T) {
		unit := &Object{Field33: 9, Field38: 7}
		new(Server).SwitchUpdate53B320(unit, SwitchUpdateRuntime53B320{})
		if unit.Field33 != 9 || unit.Field38 != 7 || !unit.ObjFlags.Has(object.FlagNoCollide) {
			t.Fatalf("state = field33:%d field38:%#x flags:%#x", unit.Field33, unit.Field38, unit.ObjFlags)
		}
	})
}

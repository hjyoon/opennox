package server

import (
	"math"
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func TestObjectSetBuffFlagsMatchesGAMEEXEContract(t *testing.T) {
	tests := []struct {
		name  string
		class object.Class
		flags uint32
	}{
		{name: "client persist", class: object.ClassClientPersist, flags: 0x10203040},
		{name: "immobile", class: object.ClassImmobile, flags: 0},
		{name: "player protection callback", class: object.ClassPlayer, flags: 0x89ABCDEF},
		{name: "ordinary flags clear", class: object.ClassMissile, flags: 0},
		{name: "ordinary flags set", class: object.ClassMissile, flags: 0x80000001},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := &Object{
				ObjClass: tc.class,
				ObjFlags: object.FlagActive,
				Buffs:    0x55555555,
				Field37:  0x80000015,
				Field38:  0x2468ACE0,
			}
			for i := range got.Field140 {
				got.Field140[i] = 0xA5000000 | 0x800000 | 0x80 | uint32(i<<4)
			}
			beforeSlots := got.Field140
			var player *Player
			if tc.class.Has(object.ClassPlayer) {
				player = &Player{ProtUnitBuffs: 0x13579BDF}
				got.UpdateData = unsafe.Pointer(&PlayerUpdateData{Player: player})
			} else if !tc.class.HasAny(object.ClassClientPersist | object.ClassImmobile) {
				bindStatusTestType(t, got, 0)
			}
			want := *got
			applyGAMEEXEBuffFlags(&want, tc.flags)

			callbackCount := 0
			got.SetBuffFlags(tc.flags, func(gotPlayer *Player, flags uint32) {
				callbackCount++
				if gotPlayer != player || flags != tc.flags {
					t.Errorf("player callback: got (%p, %#08x), want (%p, %#08x)", gotPlayer, flags, player, tc.flags)
				}
				if got.Field38 != math.MaxUint32 || got.Buffs != tc.flags || got.Field140 != beforeSlots {
					t.Error("player protection callback did not run at the original instruction boundary")
				}
			})

			wantCallbacks := 0
			if tc.class.Has(object.ClassPlayer) {
				wantCallbacks = 1
			}
			if callbackCount != wantCallbacks {
				t.Errorf("player callback count: got %d, want %d", callbackCount, wantCallbacks)
			}
			if got.Buffs != want.Buffs || got.Field38 != want.Field38 {
				t.Errorf("buffs/sync: got (%#08x, %#08x), want (%#08x, %#08x)", got.Buffs, got.Field38, want.Buffs, want.Field38)
			}
			if got.Field140 != want.Field140 {
				t.Errorf("Field140 differs: got %#v, want %#v", got.Field140, want.Field140)
			}
			if got.Field37 != want.Field37 || got.ObjClass != want.ObjClass || got.ObjFlags != want.ObjFlags || got.UpdateData != want.UpdateData {
				t.Fatal("buff update overwrote state outside the original function contract")
			}
		})
	}
}

func applyGAMEEXEBuffFlags(obj *Object, flags uint32) {
	obj.Field38 = math.MaxUint32
	obj.Buffs = flags
	if obj.Class().HasAny(object.ClassClientPersist | object.ClassImmobile | object.ClassPlayer) {
		for i := range obj.Field140 {
			obj.Field140[i] = obj.Field140[i]&0xFFFFF000 | 0x800000
		}
		return
	}
	changed := flags != 0
	for i := range obj.Field140 {
		if changed {
			obj.Field140[i] |= 0x80
		} else {
			obj.Field140[i] &^= 0x80
		}
		if obj.Field37&uint32(1<<i) != 0 {
			obj.Field140[i] |= 0x800000
		} else if obj.Field140[i]&0x80 == 0 {
			obj.Field140[i] &^= 0x800000
		}
	}
}

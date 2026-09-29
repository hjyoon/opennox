package server

import (
	"slices"
	"testing"

	"github.com/opennox/libs/object"
)

func TestScriptUnknownB8516790AndB9516850ReadExactSubclassBits(t *testing.T) {
	tests := []struct {
		subclass object.SubClass
		b8       bool
		b9       bool
	}{
		{subclass: 0, b8: false, b9: false},
		{subclass: 0x80, b8: false, b9: true},
		{subclass: 0x100, b8: true, b9: false},
		{subclass: 0x180, b8: true, b9: true},
		{subclass: 0xfffffe7f, b8: false, b9: false},
	}
	for _, tc := range tests {
		obj := &Object{ObjSubClass: tc.subclass}
		if got := ScriptUnknownB8516790(obj); got != tc.b8 {
			t.Fatalf("b8(%#x) = %v, want %v", tc.subclass, got, tc.b8)
		}
		if got := ScriptUnknownB9516850(obj); got != tc.b9 {
			t.Fatalf("b9(%#x) = %v, want %v", tc.subclass, got, tc.b9)
		}
	}
	if ScriptUnknownB8516790(nil) || ScriptUnknownB9516850(nil) {
		t.Fatal("nil object reported a subclass bit")
	}
}

func TestScriptSetHalberd516890PreservesEquippedStateAndOrder(t *testing.T) {
	host := &Object{}
	food := &Object{ObjClass: object.ClassFood}
	ordinaryWeapon := &Object{ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(object.WeaponBow)}
	old := &Object{
		ObjClass:    object.ClassWeapon,
		ObjSubClass: object.SubClass(object.WeaponStaffOblivionHeart),
		ObjFlags:    object.FlagEquipped,
	}
	host.InvFirstItem = food
	food.InvNextItem = ordinaryWeapon
	ordinaryWeapon.InvNextItem = old
	replacement := &Object{}
	var trace []string
	var deleted, equipped *Object
	var respawnOwner *Object
	var respawnType string

	ScriptSetHalberd516890(host, 2, ScriptSetHalberdRuntime516890{
		DelayedDelete: func(item *Object) {
			trace = append(trace, "delete")
			deleted = item
		},
		Respawn: func(owner *Object, typeID string) *Object {
			trace = append(trace, "respawn")
			respawnOwner, respawnType = owner, typeID
			return replacement
		},
		TryEquip: func(owner, item *Object) {
			trace = append(trace, "equip")
			if owner != host {
				t.Fatalf("equip owner = %p, want %p", owner, host)
			}
			equipped = item
		},
	})

	if want := []string{"delete", "respawn", "equip"}; !slices.Equal(trace, want) {
		t.Fatalf("trace = %v, want %v", trace, want)
	}
	if deleted != old || respawnOwner != host || respawnType != "OblivionWierdling" || equipped != replacement {
		t.Fatalf("calls = delete %p, respawn %p/%q, equip %p", deleted, respawnOwner, respawnType, equipped)
	}
}

func TestScriptSetHalberd516890TypeTableAndGuards(t *testing.T) {
	wantTypes := []string{"OblivionHalberd", "OblivionHeart", "OblivionWierdling", "OblivionOrb"}
	for upgrade, want := range wantTypes {
		var got string
		ScriptSetHalberd516890(&Object{}, upgrade, ScriptSetHalberdRuntime516890{
			DelayedDelete: func(*Object) { t.Fatal("unexpected delete") },
			Respawn: func(_ *Object, typeID string) *Object {
				got = typeID
				return nil
			},
			TryEquip: func(*Object, *Object) { t.Fatal("unexpected equip") },
		})
		if got != want {
			t.Fatalf("upgrade %d type = %q, want %q", upgrade, got, want)
		}
	}

	called := false
	runtime := ScriptSetHalberdRuntime516890{
		DelayedDelete: func(*Object) { called = true },
		Respawn:       func(*Object, string) *Object { called = true; return nil },
		TryEquip:      func(*Object, *Object) { called = true },
	}
	ScriptSetHalberd516890(nil, 0, runtime)
	ScriptSetHalberd516890(&Object{}, -1, runtime)
	ScriptSetHalberd516890(&Object{}, 4, runtime)
	if called {
		t.Fatal("invalid host or upgrade invoked runtime callbacks")
	}
}

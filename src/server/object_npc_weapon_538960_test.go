package server

import (
	"testing"

	"github.com/opennox/libs/object"
)

func TestNPCEquippedWeapon538960ReconstructsField516(t *testing.T) {
	primary := &Object{
		ObjClass:    object.ClassWand,
		ObjSubClass: object.SubClass(object.WeaponStaffFireball),
		ObjFlags:    object.FlagEquipped,
	}
	second := &Object{
		ObjClass:    object.ClassWeapon,
		ObjSubClass: object.SubClass(object.WeaponBow),
		ObjFlags:    object.FlagEquipped,
	}
	quiver := &Object{
		ObjClass:    object.ClassWeapon,
		ObjSubClass: object.SubClass(object.WeaponQuiver),
		ObjFlags:    object.FlagEquipped,
	}
	unequipped := &Object{
		ObjClass:    object.ClassWeapon,
		ObjSubClass: object.SubClass(object.WeaponCrossbow),
	}
	unrelated := &Object{
		ObjClass: object.ClassFood,
		ObjFlags: object.FlagEquipped,
	}
	unrelated.InvNextItem = unequipped
	unequipped.InvNextItem = quiver
	quiver.InvNextItem = primary
	primary.InvNextItem = second

	unit := &Object{
		ObjClass:     object.ClassMonster,
		ObjSubClass:  object.SubClass(object.MonsterNPC),
		InvFirstItem: unrelated,
	}
	if got := unit.NPCEquippedWeapon538960(); got != primary {
		t.Fatalf("equipped NPC weapon = %p, want first non-quiver primary %p", got, primary)
	}
}

func TestNPCEquippedWeapon538960RejectsIneligibleOwners(t *testing.T) {
	weapon := &Object{
		ObjClass:    object.ClassWeapon,
		ObjSubClass: object.SubClass(object.WeaponBow),
		ObjFlags:    object.FlagEquipped,
	}
	for name, unit := range map[string]*Object{
		"nil":              nil,
		"player":           {ObjClass: object.ClassPlayer, InvFirstItem: weapon},
		"ordinary monster": {ObjClass: object.ClassMonster, InvFirstItem: weapon},
		"empty NPC":        {ObjClass: object.ClassMonster, ObjSubClass: object.SubClass(object.MonsterNPC)},
	} {
		if got := unit.NPCEquippedWeapon538960(); got != nil {
			t.Fatalf("%s owner weapon = %p, want nil", name, got)
		}
	}
}

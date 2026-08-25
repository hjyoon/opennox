package server

import (
	"testing"
	"unsafe"

	"github.com/opennox/libs/object"
)

func playerActionGateUnit4F9BC0(update *PlayerUpdateData) *Object {
	return &Object{ObjClass: object.ClassPlayer, UpdateData: unsafe.Pointer(update)}
}

func TestPlayerCanMoveNative4F9BC0(t *testing.T) {
	update := &PlayerUpdateData{}
	unit := playerActionGateUnit4F9BC0(update)
	if got := playerCanMoveNative4F9BC0(unit, false); got != 1 {
		t.Fatalf("idle player can move = %d, want 1", got)
	}

	unit.Buffs = 1 << ENCHANT_FREEZE
	if got := playerCanMoveNative4F9BC0(unit, false); got != 0 {
		t.Fatalf("frozen player can move = %d, want 0", got)
	}
	unit.Buffs = 1 << ENCHANT_HELD
	if got := playerCanMoveNative4F9BC0(unit, false); got != 0 {
		t.Fatalf("held player can move = %d, want 0", got)
	}
	unit.Buffs = 0

	update.Trade70 = &TradeSession{}
	if got := playerCanMoveNative4F9BC0(unit, true); got != 0 {
		t.Fatalf("quest trader can move = %d, want 0", got)
	}
	if got := playerCanMoveNative4F9BC0(unit, false); got != 1 {
		t.Fatalf("arena trader can move = %d, want 1", got)
	}
	update.Trade70 = nil

	update.State = PlayerState1
	update.EquippedWeapon = &Object{ObjClass: object.ClassWeapon, ObjSubClass: object.SubClass(object.WeaponCrossbow)}
	if got := playerCanMoveNative4F9BC0(unit, false); got != 0 {
		t.Fatalf("crossbow attack state can move = %d, want 0", got)
	}
	update.EquippedWeapon.ObjSubClass = object.SubClass(object.WeaponBow)
	if got := playerCanMoveNative4F9BC0(unit, false); got != 1 {
		t.Fatalf("bow attack state can move = %d, want 1", got)
	}
}

func TestPlayerCanAttackNative4F9C40(t *testing.T) {
	update := &PlayerUpdateData{}
	unit := playerActionGateUnit4F9BC0(update)
	if got := playerCanAttackNative4F9C40(unit); got != 1 {
		t.Fatalf("idle player can attack = %d, want 1", got)
	}
	unit.Buffs = 1 << ENCHANT_FREEZE
	if got := playerCanAttackNative4F9C40(unit); got != 0 {
		t.Fatalf("frozen player can attack = %d, want 0", got)
	}
	unit.Buffs = 0
	update.State = PlayerState23
	if got := playerCanAttackNative4F9C40(unit); got != 0 {
		t.Fatalf("state 23 player can attack = %d, want 0", got)
	}
}

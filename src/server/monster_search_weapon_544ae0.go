package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/types"
)

type monsterSearchWeaponHooks544AE0 struct {
	eachInCircle func(types.Pointf, float32, func(*Object) bool)
	classCanUse  func(*Object, player.Class) bool
	canInteract  func(*Object, *Object, int) bool
}

type MonsterWeaponSearchState544AE0 struct {
	BestObject   *Object
	BestDistance float32
}

func monsterSearchWeapon544AE0(unit *Object, radius float32, hooks monsterSearchWeaponHooks544AE0) *Object {
	var state MonsterWeaponSearchState544AE0
	return state.search544AE0(unit, radius, hooks)
}

// GAME.EXE 00544AE0/00544B20 searches WEAPON (0x01000000), not ammunition.
// Warrior-class eligibility and interaction precede the live position reads.
// Keep the original binary32 best-distance spill and x87 C0 comparison:
// unordered distances replace the winner too, rather than being filtered.
func (state *MonsterWeaponSearchState544AE0) search544AE0(unit *Object, radius float32, hooks monsterSearchWeaponHooks544AE0) *Object {
	if unit == nil || hooks.eachInCircle == nil || hooks.classCanUse == nil || hooks.canInteract == nil {
		return nil
	}
	*state = MonsterWeaponSearchState544AE0{BestDistance: 10000000}
	hooks.eachInCircle(unit.PosVec, radius, func(candidate *Object) bool {
		state.consider544B20(candidate, unit, hooks.classCanUse, hooks.canInteract)
		return true
	})
	return state.BestObject
}

func (state *MonsterWeaponSearchState544AE0) consider544B20(candidate, unit *Object, classCanUse func(*Object, player.Class) bool,
	canInteract func(*Object, *Object, int) bool) {
	if candidate == nil || !candidate.ObjClass.Has(object.ClassWeapon) ||
		!classCanUse(candidate, player.Warrior) || !canInteract(unit, candidate, 0) {
		return
	}
	dx := monsterEdibleSubtract544A40(float64(candidate.PosVec.X), float64(unit.PosVec.X))
	dy := monsterEdibleSubtract544A40(float64(candidate.PosVec.Y), float64(unit.PosVec.Y))
	distance := monsterMoveToRunAddChop53_544434(monsterMoveToRunSquareChop53_544434(dy), monsterMoveToRunSquareChop53_544434(dx))
	if !(distance >= float64(state.BestDistance)) {
		state.BestDistance = float32(monsterMoveToRunSpill544440(distance))
		state.BestObject = candidate
	}
}

func (s *Server) MonsterSearchWeapon544AE0(unit *Object, radius float32, classCanUse func(*Object, player.Class) bool) *Object {
	return monsterSearchWeapon544AE0(unit, radius, monsterSearchWeaponHooks544AE0{
		eachInCircle: s.Map.EachObjInCircle,
		classCanUse:  classCanUse,
		canInteract:  s.CanInteract,
	})
}

// MonsterSearchWeaponNative544AE0 keeps the winner for the separate C callback.
func (s *Server) MonsterSearchWeaponNative544AE0(unit *Object, radius float32, state *MonsterWeaponSearchState544AE0,
	classCanUse func(*Object, player.Class) bool) *Object {
	return state.search544AE0(unit, radius, monsterSearchWeaponHooks544AE0{
		eachInCircle: s.Map.EachObjInCircle,
		classCanUse:  classCanUse,
		canInteract:  s.CanInteract,
	})
}

func (s *Server) MonsterSearchWeaponCandidateNative544B20(candidate, unit *Object, state *MonsterWeaponSearchState544AE0,
	classCanUse func(*Object, player.Class) bool) {
	state.consider544B20(candidate, unit, classCanUse, s.CanInteract)
}

// 00547BDE..00547C45 uses the entry-cached bot record for admission, then
// ordinary live morph/pickup/morph services. Neither the search nor inventory
// result causes an admission recheck; the reverse link is loaded after pickup.
func (s *Server) monsterMainPickupWeapon547210(unit *Object, update *MonsterUpdateData, runtime MonsterMainRuntime547210) bool {
	if !update.StatusFlags.Has(object.MonStatusBot) {
		return true
	}
	if update.Field545 == nil || update.Field545.Player == nil {
		return false // contain missing native bot metadata, not an original gate
	}
	owner := update.Field545.Player
	if owner.PlayerClass() != player.Warrior || owner.WeaponEquip != 0 || byte(s.Frame())&0xf != 0 {
		return true
	}
	if runtime.SearchWeapon == nil || runtime.PlaceInventory == nil {
		return false
	}
	weapon := runtime.SearchWeapon(unit, 75)
	if weapon != nil {
		MonsterMorphToPlayer4FAAF0(unit)
		runtime.PlaceInventory(unit, weapon, 1, 1)
		MonsterMorphFromPlayer4FAAC0(unit)
	}
	return true
}

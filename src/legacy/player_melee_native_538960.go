package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"math"
	"runtime"
	"unsafe"

	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type playerAttackDeadline538960 uint8

const (
	attackKeepDeadline538960 playerAttackDeadline538960 = iota
	attackFullDeadline538960
	attackImmediateDeadline538960
)

type meleeAnimation538960 struct {
	mask      uint32
	animation int
	missSound sound.ID
	damage    uint8
	disabled  uint32
	hammer    bool
}

var meleeAnimations538960 = [...]meleeAnimation538960{
	{0x200, 28, 880, 0, 0, false},
	{0x100, 27, 881, 0, 0, false},
	{0x400, 37, 881, 0, 0, false},
	{0x4000, 39, 882, 2, 1, true},
	{0x800, 26, 884, 2, 1, false},
	{0x3000, 35, 883, 0, 1, false},
}

type playerMeleeAttack538960 struct {
	unit     *server.Object
	weapon   *server.Object
	modifier *server.Modifier
	update   *server.PlayerUpdateData
	stored   *uint8
	previous uint8
	strength int32
}

func (a *playerMeleeAttack538960) prepare() bool {
	if a.weapon == nil {
		return true
	}
	a.modifier = GetServer().S().Modif.Nox_xxx_getProjectileClassById413250(int(a.weapon.TypeInd))
	if a.modifier == nil {
		return false
	}
	a.weapon.PosVec = a.unit.PosVec
	a.weapon.PrevPos = a.unit.PosVec
	return true
}

func (a *playerMeleeAttack538960) frame(animation int, schedule playerAttackDeadline538960) (current uint8, count int) {
	count, duration := playerAnimFrames4F9F90(animation)
	if a.update != nil && a.update.Field0 == 0 && schedule != attackKeepDeadline538960 {
		deadline := gameFrame()
		if schedule == attackFullDeadline538960 {
			deadline += uint32(count * (duration + 1))
		}
		a.update.Field0 = deadline
	}
	return uint8((gameFrame() - a.unit.Field34) / uint32(duration+1)), count
}

func (a *playerMeleeAttack538960) finish(current uint8, count int) int {
	return playerAttackProjectileFinish538960(current, count, a.stored)
}

func (a *playerMeleeAttack538960) newMiddleFrame(current uint8, count int) bool {
	return int(current) == count/2 && current > a.previous
}

func (a *playerMeleeAttack538960) ability() (result int, handled bool) {
	unit := a.unit
	raw := asObjectC(unit)
	if nox_common_playerIsAbilityActive_4FC250(raw, 2) != 0 && nox_xxx_probablyWarcryCheck_4FC3E0(raw, 2) != 0 {
		current, frames := a.frame(46, attackKeepDeadline538960)
		if a.previous == 2 && current == 3 {
			playerAttackWarcryNative538960(unit)
		}
		if int(current) >= frames {
			sub_4FC440(raw, 2)
		}
		return a.finish(current, frames), true
	}
	if nox_common_playerIsAbilityActive_4FC250(raw, 1) != 0 {
		if nox_xxx_testUnitBuffs_4FF350(raw, 25) != 0 || nox_xxx_testUnitBuffs_4FF350(raw, 5) != 0 {
			return 0, true
		}
		current, frames := a.frame(45, attackKeepDeadline538960)
		playerAttackBerserkNative538960(unit)
		if int(current) >= frames-1 {
			current = 0
		}
		return a.finish(current, frames), true
	}
	return 0, false
}

func traceMeleeAttack538960(attack server.PlayerAttackData538330) bool {
	// Modifiers receive this temporary record through legacy callbacks.
	// Keep the record in native memory and pin any Go-owned arguments.
	var pinned runtime.Pinner
	pinned.Pin(attack.Owner)
	if attack.Weapon != nil {
		pinned.Pin(attack.Weapon)
	}
	defer pinned.Unpin()
	record, free := alloc.New(server.PlayerAttackData538330{})
	defer free()
	*record = attack
	if attack.Weapon != nil && attack.Weapon.InitData != nil {
		GetServer().S().ItemApplyAttackEffect538840(attack.Weapon, attack.Owner, unsafe.Pointer(record))
	}
	var trace server.PlayerAttackTraceState538330
	return trace.Trace(attack.Owner, record, playerAttackTraceRuntime538330()) != 0
}

func (a *playerMeleeAttack538960) sound(id sound.ID) {
	GetServer().S().Audio.EventObj(id, a.unit, 0, 0)
}

func (a *playerMeleeAttack538960) punch(animation uint8) int {
	current, frames := a.frame(int(animation), attackFullDeadline538960)
	if int(current) >= frames {
		var damage server.BoltDamageModifierValues4EF1E0
		if animation >= 23 {
			damage.Minimum = 5
		}
		if animation >= 23 && animation <= 25 {
			damage.Coefficient = 0.04
		}
		if animation == 25 {
			damage.Minimum = 10
		}
		value := GetServer().S().CalcBoltDamageValues4EF1E0(a.strength, damage)
		if !traceMeleeAttack538960(server.PlayerAttackData538330{
			Damage: float32(value), DamageType: 10, Radius: a.unit.Shape.Circle.R + 20,
			Owner: a.unit, Position: a.unit.PosVec, DirectionMask: 1,
		}) {
			a.sound(879)
		}
	}
	return a.finish(current, frames)
}

func (a *playerMeleeAttack538960) hit(damageType uint8, allowDisabled uint32, miss sound.ID) {
	damage := GetServer().S().CalcBoltDamage4EF1E0(a.strength, a.modifier)
	if !traceMeleeAttack538960(server.PlayerAttackData538330{
		Damage: float32(damage), DamageType: damageType, Radius: a.unit.Shape.Circle.R + a.modifier.Range68,
		Owner: a.unit, Position: a.unit.PosVec, AllowDisabled: allowDisabled, Weapon: a.weapon, DirectionMask: 1,
	}) {
		a.sound(miss)
	}
}

func (a *playerMeleeAttack538960) hammer() {
	direction := uintptr(194136 + 8*int(int16(a.unit.Direction1)))
	damage := GetServer().S().CalcBoltDamage4EF1E0(a.strength, a.modifier)
	cosine := memmap.Float32(0x587000, direction)
	sine := memmap.Float32(0x587000, direction+4)
	traceMeleeAttack538960(server.PlayerAttackData538330{
		Damage: float32(damage), DamageType: 2, Radius: a.modifier.Range68,
		Owner: a.unit, Position: types.Ptf(a.unit.PosVec.X+float32(cosine*35), a.unit.PosVec.Y+float32(sine*35)),
		AllowDisabled: 1, Weapon: a.weapon, DirectionMask: 1,
	})
	magnitude := math.RoundToEven(float64(float32(a.strength) * 0.1))
	sendEarthquakeFromNative(&a.unit.PosVec, int(magnitude))
	a.sound(882)
}

func (a *playerMeleeAttack538960) staff(equipment uint32) int {
	data := (*server.WandUseData)(a.weapon.UseData.Ptr)
	animation := 29
	if equipment&0x8000 == 0 {
		if data == nil {
			return 0
		}
		if data.Flags&2 == 0 {
			animation = 31
		}
	}
	schedule := attackFullDeadline538960
	if animation == 31 {
		schedule = attackImmediateDeadline538960
	}
	current, frames := a.frame(animation, schedule)
	if int(current) >= frames {
		if equipment&0x047f0000 != 0 && data != nil {
			data.Flags &^= 2
		}
	} else if animation == 29 && a.newMiddleFrame(current, frames) {
		a.hit(0, 0, 879)
	} else if a.update == nil && animation == 31 && a.previous == 0 && current == 1 {
		GetServer().S().UseByNetCode53F8E0(a.unit, a.weapon)
		if data != nil && data.Charge == 0 && data.MaxCharge != 0 {
			npcWeaponDequipNative53A030(a.unit, a.weapon)
		}
	}
	return a.finish(current, frames)
}

func (a *playerMeleeAttack538960) run(animation uint8, player *server.Player, equipment uint32) int {
	a.strength = int32(a.unit.Strength())
	if result, handled := a.ability(); handled {
		return result
	}
	if a.weapon == nil {
		return a.punch(animation)
	}
	if player != nil {
		equipment = player.WeaponEquip
	}
	if equipment&0x047f8000 != 0 {
		return a.staff(equipment)
	}
	if equipment&0x07800000 != 0 {
		animation := 31
		if equipment&0x03800000 != 0 {
			animation = 32
		}
		current, frames := a.frame(animation, attackFullDeadline538960)
		if animation == 32 && a.newMiddleFrame(current, frames) {
			a.hit(0, 0, 879)
		}
		return a.finish(current, frames)
	}
	if a.update != nil && equipment&0xC0 != 0 {
		return playerAttackThrownNative538960(a.unit, a.weapon, a.update, equipment, a.previous, playerAttackThrownRuntimeFactory538960())
	}
	for _, swing := range meleeAnimations538960 {
		if equipment&swing.mask == 0 {
			continue
		}
		current, frames := a.frame(swing.animation, attackFullDeadline538960)
		if a.newMiddleFrame(current, frames) {
			if swing.hammer {
				a.hammer()
			} else {
				a.hit(swing.damage, swing.disabled, swing.missSound)
			}
		}
		return a.finish(current, frames)
	}
	if a.update == nil {
		return playerAttackNPCProjectileNative538960(a.unit, a.weapon, equipment, a.previous, a.stored, playerAttackProjectileRuntimeFactory538960())
	}
	if equipment&0xC != 0 {
		return playerAttackBowNative538960(a.unit, a.weapon, a.update, equipment, a.previous, playerAttackBowRuntimeFactory538960())
	}
	return 0
}

func playerMeleeUpdate538960(unit *server.Object, update *server.PlayerUpdateData, player *server.Player) int {
	if unit == nil || update == nil || player == nil {
		return 0
	}
	a := playerMeleeAttack538960{
		unit: unit, weapon: update.EquippedWeapon, update: update,
		previous: update.Field59_0, stored: &update.Field59_0,
	}
	animation := uint8(player.Field8)
	if !a.prepare() {
		return 0
	}
	if a.weapon == nil && animation == 0 {
		animation = uint8(nox_common_randomInt_415FA0(23, 24))
		if player.PlayerClass() == 0 && nox_common_randomInt_415FA0(0, 100) >= 75 {
			animation = 25
		}
		player.Field8 = player.Field8&0xff00 | uint16(animation)
	}
	return a.run(animation, player, 0)
}

func npcMeleeUpdate538960(unit, weapon *server.Object, equipment uint32, previous, animation uint8, stored, storedAnimation *uint8) int {
	if unit == nil || stored == nil || storedAnimation == nil {
		return 0
	}
	a := playerMeleeAttack538960{unit: unit, weapon: weapon, previous: previous, stored: stored}
	if !a.prepare() {
		return 0
	}
	if weapon == nil && animation == 0 {
		animation = uint8(nox_common_randomInt_415FA0(23, 24))
		*storedAnimation = animation
	}
	return a.run(animation, nil, equipment)
}

//export nox_xxx_playerAttackNativeData_538960
func nox_xxx_playerAttackNativeData_538960(unit *nox_object_t, update *C.nox_player_update_data_t, player *nox_playerInfo) C.int {
	return C.int(playerMeleeUpdate538960(asObjectS(unit), (*server.PlayerUpdateData)(unsafe.Pointer(update)), asPlayerS(player)))
}

//export nox_xxx_playerAttackNativeNPCData_538960
func nox_xxx_playerAttackNativeNPCData_538960(unit, weapon *nox_object_t, equipment C.uint32_t, previous, animation C.uint8_t, stored, storedAnimation *C.uint8_t) C.int {
	return C.int(npcMeleeUpdate538960(asObjectS(unit), asObjectS(weapon), uint32(equipment), uint8(previous), uint8(animation),
		(*uint8)(unsafe.Pointer(stored)), (*uint8)(unsafe.Pointer(storedAnimation))))
}

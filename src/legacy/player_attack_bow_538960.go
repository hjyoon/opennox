package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

// Player bow attacks share projectile modifiers with NPCs, but additionally
// maintain the player deadline, quiver charges, reload messages and state.
type playerAttackBowRuntime538960 struct {
	playerAttackProjectileRuntime538960
	Direction     func(server.Dir16) types.Pointf
	ReportCharges func(uint8, *server.Object, uint8, uint8)
	ReloadQuiver  func(*server.Object) int
	Priority      func(*server.Object, strman.ID, byte)
	SetState      func(*server.Object, server.PlayerState) bool
}

var playerAttackBowRuntimeFactory538960 = func() playerAttackBowRuntime538960 {
	return playerAttackBowRuntime538960{
		playerAttackProjectileRuntime538960: playerAttackProjectileRuntimeFactory538960(),
		Direction: func(direction server.Dir16) types.Pointf {
			offset := uintptr(int(int16(direction)) * 8)
			return types.Ptf(memmap.Float32(0x587000, 194136+offset), memmap.Float32(0x587000, 194140+offset))
		},
		ReportCharges: Nox_xxx_netReportCharges_4D82B0,
		ReloadQuiver: func(owner *server.Object) int {
			return int(C.nox_xxx_playerTryReloadQuiver_539FF0(asObjectC(owner)))
		},
		Priority: GetServer().S().NetPriMsgToPlayer,
		SetState: Nox_xxx_playerSetState_4FA020,
	}
}

// GAME.EXE 00539D80 caches the quiver use record before tracing/creation,
// and consumes a charge even if allocation fails after a successful trace.
func playerAttackBowShoot2_539D80(owner, quiver, weapon *server.Object, flags uint32, h playerAttackBowRuntime538960) {
	distance := float64(owner.Shape.Circle.R) + 4
	var ammo *server.AmmoUseData
	if quiver != nil {
		ammo = (*server.AmmoUseData)(quiver.UseData.Ptr)
	}
	direction := h.Direction(owner.Direction1)
	from := owner.PosVec
	dx, dy := distance*float64(direction.X), distance*float64(direction.Y)
	spawn := types.Ptf(float32(dx+float64(from.X)), float32(dy+float64(from.Y)))
	if !h.Trace(from, spawn, server.MapTraceFlags(5)) {
		return
	}
	name := "WeakArcherArrow"
	if quiver != nil {
		name = "ArcherBolt"
		if flags == uint32(object.WeaponBow) {
			name = "ArcherArrow"
		}
	}
	if projectile := h.NewObject(name); projectile != nil {
		(*server.ArrowCollideData)(projectile.CollideData).Owner = owner
		h.CreateAt(projectile, owner, spawn)
		if quiver != nil {
			h.ApplyModifierAttrs(projectile, (*server.ModifierInitData)(quiver.InitData))
		}
		playerAttackProjectileApplyEffects539F40(owner, weapon, projectile, h.playerAttackProjectileRuntime538960)
		projectile.VelVec.X = h.Direction(owner.Direction1).X * projectile.SpeedCur
		projectile.VelVec.Y = h.Direction(owner.Direction1).Y * projectile.SpeedCur
		projectile.Direction1, projectile.Direction2 = owner.Direction1, owner.Direction1
	}
	if quiver != nil && ammo.Field2 == 0 && owner.Class().Has(object.ClassPlayer) {
		update := owner.UpdateDataPlayer()
		ammo.Charge1--
		h.ReportCharges(update.Player.PlayerInd, quiver, ammo.Charge1, ammo.Charge0)
		if ammo.Charge1 == 0 {
			h.DelayedDelete(quiver)
		}
	}
	id := sound.ID(886)
	if flags == uint32(object.WeaponBow) {
		id = 885
	}
	h.AudioEvent(id, owner)
}

// GAME.EXE 00539BD0 distinguishes a usable equipped quiver from a new
// quiver to auto-equip, and retains the Quest weak-arrow fallback.
func playerAttackBowShoot1_539BD0(owner, weapon *server.Object, h playerAttackBowRuntime538960) int {
	flags := h.WeaponInventoryFlags(weapon)
	ammo := (*server.AmmoUseData)(weapon.UseData.Ptr)
	if ammo.Charge0 != 0 {
		h.Priority(owner, "pattack.c:ReloadingQuiver", 0)
		if flags == uint32(object.WeaponBow) {
			if !h.QuestMode() {
				h.AudioEvent(887, owner)
			}
		} else if flags == uint32(object.WeaponCrossbow) {
			h.AudioEvent(888, owner)
		}
		ammo.Charge0--
		return 0
	}
	for quiver := owner.InvFirstItem; quiver != nil; quiver = quiver.InvNextItem {
		if !quiver.Flags().Has(object.FlagEquipped) || h.WeaponInventoryFlags(quiver) != 2 {
			continue
		}
		use := (*server.AmmoUseData)(quiver.UseData.Ptr)
		if use.Charge1 != 0 || use.Field2 == 1 {
			playerAttackBowShoot2_539D80(owner, quiver, weapon, flags, h)
			return 1
		}
	}
	if h.QuestMode() && flags == uint32(object.WeaponBow) {
		playerAttackBowShoot2_539D80(owner, nil, weapon, flags, h)
	}
	if !owner.Class().Has(object.ClassPlayer) {
		return 0
	}
	if h.ReloadQuiver(owner) == 1 {
		h.Priority(owner, "pattack.c:ReloadQuiver", 0)
		ammo.Charge0 = 0
	} else if !h.QuestMode() || flags != uint32(object.WeaponBow) {
		h.Priority(owner, "pattack.c:NoQuiver", 0)
	}
	if flags == uint32(object.WeaponBow) {
		if !h.QuestMode() {
			h.AudioEvent(887, owner)
		}
	} else if flags == uint32(object.WeaponCrossbow) {
		h.AudioEvent(888, owner)
	}
	h.SetState(owner, server.PlayerState13)
	return 0
}

func playerAttackBowNative538960(owner, weapon *server.Object, update *server.PlayerUpdateData, equipment uint32, previous uint8, h playerAttackBowRuntime538960) int {
	action := 34
	if equipment&uint32(object.WeaponBow) != 0 {
		action = 33
	}
	frames, duration := h.AnimFrames(action)
	step := uint32(duration) + 1
	if owner.Class().Has(object.ClassPlayer) && update.Field0 == 0 {
		readiness := h.Readiness(weapon)
		update.Field0 = h.Frame() + uint32(frames)*step - uint32(readiness)
	}
	var current uint8
	if action == 33 {
		base := (h.Frame() - owner.Field34) / step
		adjusted := int32(base) + h.Readiness(weapon)
		current = uint8(adjusted)
		if adjusted >= int32(frames)-1 && adjusted > int32(previous) {
			playerAttackBowShoot1_539BD0(owner, weapon, h)
			current = uint8(frames)
		}
	} else {
		current = uint8((h.Frame() - owner.Field34) / step)
		if current == 1 && previous == 0 {
			playerAttackBowShoot1_539BD0(owner, weapon, h)
		}
		if current >= 1 {
			current += uint8(h.Readiness(weapon))
		}
	}
	return playerAttackThrownFinish538960(owner, update, current, frames)
}

//export nox_xxx_playerAttackBowNative_538960
func nox_xxx_playerAttackBowNative_538960(cowner, cweapon *nox_object_t, cupdate *C.nox_player_update_data_t, equipment C.uint32_t, previous C.uint8_t) C.int {
	return C.int(playerAttackBowNative538960(asObjectS(cowner), asObjectS(cweapon),
		(*server.PlayerUpdateData)(unsafe.Pointer(cupdate)), uint32(equipment), uint8(previous), playerAttackBowRuntimeFactory538960()))
}

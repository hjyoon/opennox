package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

// playerAttackProjectileRuntime538960 keeps the recovered attack sequencing
// independent from every engine service that creates, links, or destroys an
// Object. Tests can therefore exercise the C -> Go boundary with high native
// pointers without depending on the live object-type registry.
type playerAttackProjectileRuntime538960 struct {
	Frame                 func() uint32
	AnimFrames            func(int) (int, int)
	Readiness             func(*server.Object) int32
	Trace                 func(types.Pointf, types.Pointf, server.MapTraceFlags) bool
	NewObject             func(string) *server.Object
	CreateAt              func(*server.Object, *server.Object, types.Pointf)
	ApplyModifierAttrs    func(*server.Object, *server.ModifierInitData)
	DetachInventory       func(*server.Object, *server.Object)
	InventoryPut          func(*server.Object, *server.Object, bool)
	AudioEvent            func(sound.ID, *server.Object)
	DelayedDelete         func(*server.Object)
	WeaponInventoryFlags  func(*server.Object) uint32
	QuestMode             func() bool
	EquipWeapon           func(*server.Object, *server.Object) int
	RecoilEffect          unsafe.Pointer
	ProjectileSpeedEffect unsafe.Pointer
	ApplyProjectileSpeed  func(*server.ModifierEff, *server.Object)
}

var playerAttackProjectileRuntimeFactory538960 = func() playerAttackProjectileRuntime538960 {
	outer := GetServer()
	srv := outer.S()
	return playerAttackProjectileRuntime538960{
		Frame:      srv.Frame,
		AnimFrames: playerAnimFrames4F9F90,
		Readiness:  Nox_xxx_itemCheckReadinessEffect_4E0960,
		Trace:      srv.MapTraceRay,
		NewObject:  srv.NewObjectByTypeID,
		CreateAt: func(obj, owner *server.Object, pos types.Pointf) {
			outer.CreateObjectAt(obj, owner, pos)
		},
		ApplyModifierAttrs: func(obj *server.Object, attrs *server.ModifierInitData) {
			srv.ApplyModifierAttrs4E4990(obj, attrs)
		},
		DetachInventory: inventoryDetach4ED0C0,
		InventoryPut:    inventoryPutImpl4F3070,
		AudioEvent: func(id sound.ID, obj *server.Object) {
			srv.Audio.EventObj(id, obj, 0, 0)
		},
		DelayedDelete:        outer.DelayedDelete,
		WeaponInventoryFlags: objectNPCWeaponEquipFlags,
		QuestMode: func() bool {
			return noxflags.HasGame(noxflags.GameModeQuest)
		},
		EquipWeapon:           npcWeaponEquipNative53A2C0,
		RecoilEffect:          RecoilEffectPointer4E0640(),
		ProjectileSpeedEffect: ProjectileSpeedEffectPointer4E09B0(),
		ApplyProjectileSpeed: func(effect *server.ModifierEff, projectile *server.Object) {
			modifierProjectileSpeedCallNative4E09B0(effect, projectile)
		},
	}
}

func playerAttackProjectileFrame538960(
	unit *server.Object, action int, runtime playerAttackProjectileRuntime538960,
) (uint8, int, bool) {
	if unit == nil || runtime.Frame == nil || runtime.AnimFrames == nil {
		return 0, 0, false
	}
	frames, duration := runtime.AnimFrames(action)
	if duration < 0 {
		return 0, frames, false
	}
	current := uint8((runtime.Frame() - unit.Field34) / uint32(duration+1))
	return current, frames, true
}

func playerAttackProjectileFinish538960(current uint8, frames int, stored *uint8) int {
	if stored != nil {
		*stored = current
		if int(current) >= frames {
			*stored = uint8(frames - 1)
		}
	}
	if int(current) < frames {
		return 1
	}
	return 0
}

func playerAttackProjectileSpawnPoint538960(unit *server.Object) types.Pointf {
	distance := unit.Shape.Circle.R + 4
	cosine, sine := server.SinCosDir(byte(unit.Direction1))
	return types.Pointf{
		X: unit.PosVec.X + distance*cosine,
		Y: unit.PosVec.Y + distance*sine,
	}
}

func playerAttackProjectileMotion538960(unit, projectile *server.Object) {
	cosine, sine := server.SinCosDir(byte(unit.Direction1))
	projectile.VelVec.X = cosine * projectile.SpeedCur
	projectile.VelVec.Y = sine * projectile.SpeedCur
	projectile.Direction1 = unit.Direction1
	projectile.Direction2 = unit.Direction1
}

func playerAttackProjectileApplyEffects539F40(
	owner, weapon, projectile *server.Object, runtime playerAttackProjectileRuntime538960,
) {
	if owner == nil || weapon == nil || projectile == nil ||
		weapon.InitData == nil || projectile.InitData == nil {
		return
	}
	weaponData := (*server.ModifierInitData)(weapon.InitData)
	projectileData := (*server.ModifierInitData)(projectile.InitData)
	for _, effect := range weaponData.Modifiers[2:] {
		if effect == nil {
			continue
		}
		if effect.AttackPreHit52.Fnc == runtime.RecoilEffect {
			projectileData.Modifiers[3] = effect
		} else if effect.Attack40.Fnc == runtime.ProjectileSpeedEffect &&
			runtime.ApplyProjectileSpeed != nil {
			runtime.ApplyProjectileSpeed(effect, projectile)
		}
	}
}

func playerAttackShootBowCrossbow2_539D80(
	owner, quiver, weapon *server.Object, weaponFlags uint32,
	runtime playerAttackProjectileRuntime538960,
) bool {
	spawn := playerAttackProjectileSpawnPoint538960(owner)
	if runtime.Trace == nil || !runtime.Trace(owner.PosVec, spawn, server.MapTraceFlags(5)) {
		return false
	}

	projectileType := "WeakArcherArrow"
	if quiver != nil {
		if weaponFlags == uint32(object.WeaponBow) {
			projectileType = "ArcherArrow"
		} else {
			projectileType = "ArcherBolt"
		}
	}
	var projectile *server.Object
	if runtime.NewObject != nil {
		projectile = runtime.NewObject(projectileType)
	}
	if projectile != nil {
		if data := (*server.ArrowCollideData)(projectile.CollideData); data != nil {
			data.Owner = owner
		}
		if runtime.CreateAt != nil {
			runtime.CreateAt(projectile, owner, spawn)
		}
		if quiver != nil && quiver.InitData != nil && runtime.ApplyModifierAttrs != nil {
			runtime.ApplyModifierAttrs(projectile, (*server.ModifierInitData)(quiver.InitData))
		}
		playerAttackProjectileApplyEffects539F40(owner, weapon, projectile, runtime)
		playerAttackProjectileMotion538960(owner, projectile)
	}
	if runtime.AudioEvent != nil {
		if weaponFlags == uint32(object.WeaponBow) {
			runtime.AudioEvent(sound.ID(885), owner)
		} else {
			runtime.AudioEvent(sound.ID(886), owner)
		}
	}
	return true
}

func playerAttackShootBowCrossbow1_539BD0(
	owner, weapon *server.Object, runtime playerAttackProjectileRuntime538960,
) int {
	if owner == nil || weapon == nil || runtime.WeaponInventoryFlags == nil {
		return 0
	}
	weaponFlags := runtime.WeaponInventoryFlags(weapon)
	weaponAmmo := (*server.AmmoUseData)(weapon.UseData.Ptr)
	if weaponAmmo == nil {
		return 0
	}
	if weaponAmmo.Charge0 == 0 {
		for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
			if !item.Flags().Has(object.FlagEquipped) ||
				runtime.WeaponInventoryFlags(item) != uint32(object.WeaponQuiver) {
				continue
			}
			ammo := (*server.AmmoUseData)(item.UseData.Ptr)
			if ammo == nil || ammo.Charge1 == 0 && ammo.Field2 != 1 {
				continue
			}
			playerAttackShootBowCrossbow2_539D80(owner, item, weapon, weaponFlags, runtime)
			return 1
		}
		if runtime.QuestMode != nil && runtime.QuestMode() &&
			weaponFlags == uint32(object.WeaponBow) {
			playerAttackShootBowCrossbow2_539D80(owner, nil, weapon, weaponFlags, runtime)
		}
		// NPCs do not execute the player-only reload messages, state change, or
		// missing-ammunition sounds that follow this point in GAME.EXE.
		return 0
	}

	if runtime.AudioEvent != nil {
		switch weaponFlags {
		case uint32(object.WeaponBow):
			if runtime.QuestMode == nil || !runtime.QuestMode() {
				runtime.AudioEvent(sound.ID(887), owner)
			}
		case uint32(object.WeaponCrossbow):
			runtime.AudioEvent(sound.ID(888), owner)
		}
	}
	weaponAmmo.Charge0--
	return 0
}

func playerAttackRoundChakram538960(
	owner, weapon *server.Object, previous uint8, stored *uint8,
	runtime playerAttackProjectileRuntime538960,
) int {
	current, frames, ok := playerAttackProjectileFrame538960(owner, 44, runtime)
	if !ok {
		return 0
	}
	if current == uint8(frames/2) && current > previous {
		spawn := playerAttackProjectileSpawnPoint538960(owner)
		if runtime.Trace == nil || !runtime.Trace(owner.PosVec, spawn, server.MapTraceFlags(5)) {
			if runtime.AudioEvent != nil {
				runtime.AudioEvent(sound.ID(323), owner)
			}
			return playerAttackProjectileFinish538960(current, frames, stored)
		}
		var projectile *server.Object
		if runtime.NewObject != nil {
			projectile = runtime.NewObject("RoundChakramInMotion")
		}
		if projectile == nil {
			return 0
		}
		data := (*server.ChakramUpdateData)(projectile.UpdateData)
		if data == nil {
			if runtime.DelayedDelete != nil {
				runtime.DelayedDelete(projectile)
			}
			return 0
		}
		if runtime.DetachInventory != nil {
			runtime.DetachInventory(owner, weapon)
		}
		if runtime.CreateAt != nil {
			runtime.CreateAt(projectile, owner, spawn)
		}
		if runtime.InventoryPut != nil {
			runtime.InventoryPut(projectile, weapon, true)
		}
		if weapon.InitData != nil && runtime.ApplyModifierAttrs != nil {
			runtime.ApplyModifierAttrs(projectile, (*server.ModifierInitData)(weapon.InitData))
		}
		playerAttackProjectileMotion538960(owner, projectile)
		data.Reflections = 4
		data.OwnerPos = owner.PosVec
		data.ReturnState = 2
		if runtime.AudioEvent != nil {
			runtime.AudioEvent(sound.ID(891), owner)
		}
	}
	return playerAttackProjectileFinish538960(current, frames, stored)
}

func playerAttackFanChakram538960(
	owner, weapon *server.Object, previous uint8, stored *uint8,
	runtime playerAttackProjectileRuntime538960,
) int {
	current, frames, ok := playerAttackProjectileFrame538960(owner, 44, runtime)
	if !ok {
		return 0
	}
	if current == uint8(frames/2) && current > previous {
		spawn := playerAttackProjectileSpawnPoint538960(owner)
		if runtime.Trace == nil || !runtime.Trace(owner.PosVec, spawn, server.MapTraceFlags(4)) {
			if runtime.AudioEvent != nil {
				runtime.AudioEvent(sound.ID(323), owner)
			}
			return playerAttackProjectileFinish538960(current, frames, stored)
		}
		var projectile *server.Object
		if runtime.NewObject != nil {
			projectile = runtime.NewObject("FanChakramInMotion")
		}
		if projectile == nil {
			return 0
		}
		if data := (*server.ArrowCollideData)(projectile.CollideData); data != nil {
			data.Owner = owner
		}
		if runtime.CreateAt != nil {
			runtime.CreateAt(projectile, owner, spawn)
		}
		if weapon.InitData != nil && runtime.ApplyModifierAttrs != nil {
			runtime.ApplyModifierAttrs(projectile, (*server.ModifierInitData)(weapon.InitData))
		}
		playerAttackProjectileMotion538960(owner, projectile)
		if runtime.AudioEvent != nil {
			runtime.AudioEvent(sound.ID(891), owner)
		}

		ammo := (*server.AmmoUseData)(weapon.UseData.Ptr)
		if ammo != nil && ammo.Field2 == 0 {
			ammo.Charge1--
			if ammo.Charge1 == 0 {
				if runtime.DetachInventory != nil {
					runtime.DetachInventory(owner, weapon)
				}
				if runtime.DelayedDelete != nil {
					runtime.DelayedDelete(weapon)
				}
				if runtime.WeaponInventoryFlags != nil && runtime.EquipWeapon != nil {
					for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
						if runtime.WeaponInventoryFlags(item) == uint32(object.WeaponShuriken) {
							runtime.EquipWeapon(owner, item)
							break
						}
					}
				}
			}
		}
	}
	return playerAttackProjectileFinish538960(current, frames, stored)
}

func playerAttackBowCrossbow538960(
	owner, weapon *server.Object, equipment uint32, previous uint8, stored *uint8,
	runtime playerAttackProjectileRuntime538960,
) int {
	action := 33
	if equipment&uint32(object.WeaponCrossbow) != 0 {
		action = 34
	}
	current, frames, ok := playerAttackProjectileFrame538960(owner, action, runtime)
	if !ok {
		return 0
	}
	if action == 33 {
		adjusted := int(current)
		if runtime.Readiness != nil {
			adjusted += int(runtime.Readiness(weapon))
		}
		current = uint8(adjusted)
		if adjusted >= frames-1 && adjusted > int(previous) {
			playerAttackShootBowCrossbow1_539BD0(owner, weapon, runtime)
			current = uint8(frames)
		}
	} else {
		if current == 1 && previous == 0 {
			playerAttackShootBowCrossbow1_539BD0(owner, weapon, runtime)
		}
		if current >= 1 && runtime.Readiness != nil {
			current += uint8(runtime.Readiness(weapon))
		}
	}
	return playerAttackProjectileFinish538960(current, frames, stored)
}

func playerAttackNPCProjectileNative538960(
	owner, weapon *server.Object, equipment uint32, previous uint8, stored *uint8,
	runtime playerAttackProjectileRuntime538960,
) int {
	if owner == nil || weapon == nil || stored == nil {
		return 0
	}
	switch {
	case equipment&uint32(object.WeaponChakram) != 0:
		return playerAttackRoundChakram538960(owner, weapon, previous, stored, runtime)
	case equipment&uint32(object.WeaponShuriken) != 0:
		return playerAttackFanChakram538960(owner, weapon, previous, stored, runtime)
	case equipment&uint32(object.WeaponBow|object.WeaponCrossbow) != 0:
		return playerAttackBowCrossbow538960(owner, weapon, equipment, previous, stored, runtime)
	default:
		return 0
	}
}

//export nox_xxx_playerAttackNPCProjectileNative_538960
func nox_xxx_playerAttackNPCProjectileNative_538960(
	cowner, cweapon *nox_object_t, equipment C.uint32_t,
	previous C.uint8_t, stored *C.uint8_t,
) C.int {
	return C.int(playerAttackNPCProjectileNative538960(
		asObjectS(cowner), asObjectS(cweapon), uint32(equipment), uint8(previous),
		(*uint8)(unsafe.Pointer(stored)), playerAttackProjectileRuntimeFactory538960(),
	))
}

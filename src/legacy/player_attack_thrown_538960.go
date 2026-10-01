package legacy

/*
#include "GAME4_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

// These are the player-only branches of GAME.EXE 00538960 (00539005 and
// 0053918E), not the NPC adapter: the player deadline, charge packet, and
// final-shuriken auto-equip must reach their original engine services.
type playerAttackThrownRuntime538960 struct {
	Frame              func() uint32
	AnimFrames         func(int) (int, int)
	Direction          func(server.Dir16) types.Pointf
	Trace              func(types.Pointf, types.Pointf, server.MapTraceFlags) bool
	NewObject          func(string) *server.Object
	CreateAt           func(*server.Object, *server.Object, types.Pointf)
	ApplyModifierAttrs func(*server.Object, *server.ModifierInitData)
	DetachInventory    func(*server.Object, *server.Object)
	InventoryPut       func(*server.Object, *server.Object, bool)
	AudioEvent         func(sound.ID, *server.Object)
	DelayedDelete      func(*server.Object)
	ReportCharges      func(uint8, *server.Object, uint8, uint8)
	AutoEquipShuriken  func(*server.Object)
}

var playerAttackThrownRuntimeFactory538960 = func() playerAttackThrownRuntime538960 {
	outer := GetServer()
	srv := outer.S()
	return playerAttackThrownRuntime538960{
		Frame:      srv.Frame,
		AnimFrames: playerAnimFrames4F9F90,
		Direction: func(direction server.Dir16) types.Pointf {
			// The original table index sign-extends the full direction WORD;
			// it does not truncate to a byte or substitute a computed angle.
			offset := uintptr(int(int16(direction)) * 8)
			return types.Ptf(memmap.Float32(0x587000, 194136+offset), memmap.Float32(0x587000, 194140+offset))
		},
		Trace:     srv.MapTraceRay,
		NewObject: srv.NewObjectByTypeID,
		CreateAt: func(projectile, owner *server.Object, pos types.Pointf) {
			outer.CreateObjectAt(projectile, owner, pos)
		},
		ApplyModifierAttrs: func(projectile *server.Object, attrs *server.ModifierInitData) {
			srv.ApplyModifierAttrs4E4990(projectile, attrs)
		},
		DetachInventory: inventoryDetach4ED0C0,
		InventoryPut:    inventoryPutImpl4F3070,
		AudioEvent: func(id sound.ID, owner *server.Object) {
			srv.Audio.EventObj(id, owner, 0, 0)
		},
		DelayedDelete: outer.DelayedDelete,
		ReportCharges: Nox_xxx_netReportCharges_4D82B0,
		AutoEquipShuriken: func(owner *server.Object) {
			C.sub_539FB0(asObjectC(owner))
		},
	}
}

func playerAttackThrownFinish538960(
	owner *server.Object, update *server.PlayerUpdateData, current uint8, frames int,
) int {
	var stored *uint8
	if owner.Class().Has(object.ClassPlayer) {
		stored = &update.Field59_0
	} else if owner.Class().Has(object.ClassMonster) && owner.MonsterClass().Has(object.MonsterNPC) {
		// As in 00539AE3, the cached update pointer survives callbacks, but
		// the live owner class chooses which one-byte frame slot is stored.
		stored = (*uint8)(unsafe.Pointer(&(*server.MonsterUpdateData)(unsafe.Pointer(update)).Field481))
	}
	return playerAttackProjectileFinish538960(current, frames, stored)
}

func playerAttackThrownNative538960(
	owner, weapon *server.Object, update *server.PlayerUpdateData,
	equipment uint32, previous uint8, deps playerAttackThrownRuntime538960,
) int {
	frames, duration := deps.AnimFrames(44)
	step := uint32(duration) + 1
	if owner.Class().Has(object.ClassPlayer) && update.Field0 == 0 {
		update.Field0 = deps.Frame() + uint32(frames)*step
	}
	current := uint8((deps.Frame() - owner.Field34) / step)
	if int(current) != frames/2 || current <= previous {
		return playerAttackThrownFinish538960(owner, update, current, frames)
	}

	// x87 retains radius+4 and each product until the trace's binary32
	// coordinate stores. Do not add the earlier float32 spills of the NPC
	// helper, or recompute this position after creation/inventory callbacks.
	distance := float64(float64(owner.Shape.Circle.R) + 4)
	direction := deps.Direction(owner.Direction1)
	from := owner.PosVec
	dx := float64(distance * float64(direction.X))
	dy := float64(distance * float64(direction.Y))
	spawn := types.Ptf(float32(dx+float64(from.X)), float32(dy+float64(from.Y)))

	// Round has priority over Fan when both equipment bits are present.
	flags := server.MapTraceFlags(5)
	name := "RoundChakramInMotion"
	var ammo *server.AmmoUseData
	if equipment&uint32(object.WeaponChakram) == 0 {
		flags, name = 4, "FanChakramInMotion"
		ammo = (*server.AmmoUseData)(weapon.UseData.Ptr) // cached before trace
	}
	if !deps.Trace(from, spawn, flags) {
		deps.AudioEvent(sound.ID(323), owner)
		return playerAttackThrownFinish538960(owner, update, current, frames)
	}
	projectile := deps.NewObject(name)
	if projectile == nil {
		return 0 // the original allocation failure does not store the frame
	}
	var chakram *server.ChakramUpdateData
	if equipment&uint32(object.WeaponChakram) != 0 {
		chakram = (*server.ChakramUpdateData)(projectile.UpdateData) // cached before detach
		deps.DetachInventory(owner, weapon)
	} else {
		(*server.ArrowCollideData)(projectile.CollideData).Owner = owner
	}
	deps.CreateAt(projectile, owner, spawn)
	if equipment&uint32(object.WeaponChakram) != 0 {
		deps.InventoryPut(projectile, weapon, true)
	}
	deps.ApplyModifierAttrs(projectile, (*server.ModifierInitData)(weapon.InitData))
	direction = deps.Direction(owner.Direction1)
	projectile.VelVec.X = direction.X * projectile.SpeedCur
	projectile.VelVec.Y = direction.Y * projectile.SpeedCur
	projectile.Direction1 = owner.Direction1
	projectile.Direction2 = owner.Direction1
	if equipment&uint32(object.WeaponChakram) != 0 {
		chakram.Reflections = 4
		chakram.OwnerPos = owner.PosVec
		chakram.ReturnState = 2
	}
	deps.AudioEvent(sound.ID(891), owner)
	if equipment&uint32(object.WeaponChakram) == 0 {
		if ammo.Field2 == 0 {
			ammo.Charge1--
			if ammo.Charge1 == 0 {
				deps.DetachInventory(owner, weapon)
				deps.DelayedDelete(weapon)
				deps.AutoEquipShuriken(owner)
			} else if owner.Class().Has(object.ClassPlayer) {
				deps.ReportCharges(update.Player.PlayerInd, weapon, ammo.Charge1, ammo.Charge0)
			}
		}
	}
	return playerAttackThrownFinish538960(owner, update, current, frames)
}

//export nox_xxx_playerAttackThrownNative_538960
func nox_xxx_playerAttackThrownNative_538960(
	cowner, cweapon *nox_object_t, cupdate *C.nox_player_update_data_t,
	equipment C.uint32_t, previous C.uint8_t,
) C.int {
	return C.int(playerAttackThrownNative538960(
		asObjectS(cowner), asObjectS(cweapon), (*server.PlayerUpdateData)(unsafe.Pointer(cupdate)),
		uint32(equipment), uint8(previous), playerAttackThrownRuntimeFactory538960(),
	))
}

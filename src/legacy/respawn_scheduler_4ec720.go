package legacy

/*
#include <stdint.h>

extern uint32_t dword_5d4594_1568028;

static uint32_t nox_respawn_scheduler_load_crown_4EC720(void) {
	return dword_5d4594_1568028;
}

static void nox_respawn_scheduler_store_crown_4EC720(uint32_t value) {
	dword_5d4594_1568028 = value;
}
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

// Keep the historical local name for the scheduler tests while sharing the
// canonical fixed-width WandUse/WandCastUse record with WeaponXfer.
type respawnRechargeData4EC720 = server.WandUseData

// respawnRechargeItem4EC720 is the pointer-width-independent specialization
// of nox_xxx_rechargeItem_53C520 used by the scheduler. Its return value is
// retained for differential tests even though sub_4EC720 ignores it.
func respawnRechargeItem4EC720(obj *server.Object, amount uint32) bool {
	data := (*respawnRechargeData4EC720)(obj.UseData.Ptr)
	if data == nil {
		return false
	}
	progress := int32(data.Progress)
	if progress >= 100 {
		return false
	}
	progress += int32(amount)
	if progress < 100 {
		data.Progress = uint32(progress)
	} else {
		data.Progress = 100
	}
	charge := int32(data.Progress) * int32(data.MaxCharge) / 100
	if charge == int32(data.Charge) {
		return false
	}
	data.Charge = uint8(charge)
	return true
}

type respawnSchedulerRuntime4EC720 struct {
	LookupTypeByName   func(string) uint32
	GameFlagsCheck     func(uint32) uint32
	UnitDefAllowed     func(uint32) bool
	NewObjectByTypeInd func(uint32) *server.Object
	SpecialModeCheck   func(uint32) uint32
	FPS                func() uint32
	Frame              func() uint32
	PointFX            func(uint32, types.Pointf)
	AudioAtPosition    func(uint32, types.Pointf)
	AudioOnObject      func(uint32, *server.Object)
	MoveTo             func(*server.Object, types.Pointf)
	WeaponEquipFlags   func(*server.Object) uint32
	SetHP              func(*server.Object, uint16)
	CreateAt           func(*server.Object, *server.Object, types.Pointf)
	CopyModifierAttrs  func(*server.Object, *server.ModifierInitData)
	DelayedDelete      func(*server.Object)
}

func respawnSchedulerLoadCrown4EC720() uint32 {
	return uint32(C.nox_respawn_scheduler_load_crown_4EC720())
}

func respawnSchedulerStoreCrown4EC720(value uint32) {
	C.nox_respawn_scheduler_store_crown_4EC720(C.uint32_t(value))
}

func respawnSchedulerNative4EC720(runtime respawnSchedulerRuntime4EC720) {
	server.RespawnScheduler4EC720(server.RespawnSchedulerHooks4EC720[
		*server.Object, *respawnRecord4EC5E0, unsafe.Pointer, *server.HealthData,
	]{
		LoadCrown:        respawnSchedulerLoadCrown4EC720,
		LookupTypeByName: runtime.LookupTypeByName,
		StoreCrown:       respawnSchedulerStoreCrown4EC720,
		GameFlagsCheck:   runtime.GameFlagsCheck,
		LoadHead:         respawnAddLoadHead4EC5E0,
		StoreAllow:       respawnAddStoreAllow4EC5E0,
		LoadPending: func(rec *respawnRecord4EC5E0) uint32 {
			return rec.Pending
		},
		StorePending: func(rec *respawnRecord4EC5E0, value uint32) {
			rec.Pending = value
		},
		LoadRespawnAt: func(rec *respawnRecord4EC5E0) uint32 {
			return rec.RespawnAt
		},
		StoreRespawnAt: func(rec *respawnRecord4EC5E0, value uint32) {
			rec.RespawnAt = value
		},
		LoadObject: func(rec *respawnRecord4EC5E0) *server.Object {
			return rec.Object
		},
		StoreObject: func(rec *respawnRecord4EC5E0, obj *server.Object) {
			rec.Object = obj
		},
		LoadNext: func(rec *respawnRecord4EC5E0) *respawnRecord4EC5E0 {
			return rec.Next
		},
		LoadRecordTypeInd: func(rec *respawnRecord4EC5E0) uint32 {
			return rec.TypeInd
		},
		LoadRecordX: func(rec *respawnRecord4EC5E0) float32 {
			return rec.X
		},
		LoadRecordY: func(rec *respawnRecord4EC5E0) float32 {
			return rec.Y
		},
		LoadDirection: func(rec *respawnRecord4EC5E0) uint16 {
			return rec.Direction
		},
		LoadCharge1: func(rec *respawnRecord4EC5E0) uint8 {
			return rec.Charge1
		},
		LoadCharge0: func(rec *respawnRecord4EC5E0) uint8 {
			return rec.Charge0
		},
		LoadObjectTypeInd: func(obj *server.Object) uint16 {
			return obj.TypeInd
		},
		LoadClass: func(obj *server.Object) uint32 {
			return uint32(obj.ObjClass)
		},
		LoadFlags: func(obj *server.Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		LoadInvHolder: func(obj *server.Object) *server.Object {
			return obj.InvHolder
		},
		LoadField32: func(obj *server.Object) uint32 {
			return obj.Field32
		},
		LoadObjectX: func(obj *server.Object) float32 {
			return obj.PosVec.X
		},
		LoadObjectY: func(obj *server.Object) float32 {
			return obj.PosVec.Y
		},
		LoadUseData: func(obj *server.Object) unsafe.Pointer {
			return obj.UseData.Ptr
		},
		StoreUseByte: func(data unsafe.Pointer, index uint32, value uint8) {
			*(*uint8)(unsafe.Add(data, uintptr(index))) = value
		},
		LoadHealthData: func(obj *server.Object) *server.HealthData {
			return obj.HealthData
		},
		LoadHealthMax: func(health *server.HealthData) uint16 {
			return health.Max
		},
		UnitDefAllowed:     runtime.UnitDefAllowed,
		NewObjectByTypeInd: runtime.NewObjectByTypeInd,
		SpecialModeCheck:   runtime.SpecialModeCheck,
		LoadFPS:            runtime.FPS,
		LoadFrame:          runtime.Frame,
		PointFXAtObject: func(code uint32, obj *server.Object) {
			runtime.PointFX(code, obj.PosVec)
		},
		PointFXAtRecord: func(code uint32, rec *respawnRecord4EC5E0) {
			runtime.PointFX(code, types.Pointf{X: rec.X, Y: rec.Y})
		},
		AudioAtObjectPosition: func(code uint32, obj *server.Object) {
			runtime.AudioAtPosition(code, obj.PosVec)
		},
		AudioAtRecordPosition: func(code uint32, rec *respawnRecord4EC5E0) {
			runtime.AudioAtPosition(code, types.Pointf{X: rec.X, Y: rec.Y})
		},
		AudioOnObject: runtime.AudioOnObject,
		MoveToRecord: func(obj *server.Object, rec *respawnRecord4EC5E0) {
			runtime.MoveTo(obj, types.Pointf{X: rec.X, Y: rec.Y})
		},
		Recharge: func(obj *server.Object, amount uint32) {
			respawnRechargeItem4EC720(obj, amount)
		},
		WeaponEquipFlags: runtime.WeaponEquipFlags,
		SetHP:            runtime.SetHP,
		CreateAt: func(obj, owner *server.Object, x, y float32) {
			runtime.CreateAt(obj, owner, types.Pointf{X: x, Y: y})
		},
		StoreDirection1: func(obj *server.Object, value uint16) {
			obj.Direction1 = server.Dir16(value)
		},
		StoreDirection2: func(obj *server.Object, value uint16) {
			obj.Direction2 = server.Dir16(value)
		},
		CopyModifierAttrs: func(obj *server.Object, rec *respawnRecord4EC5E0) {
			runtime.CopyModifierAttrs(obj, &rec.Attrs)
		},
		DelayedDelete: runtime.DelayedDelete,
	})
}

// Sub_4EC720 restores GAME.EXE 004EC720 using the native-width respawn list
// and pointer-width-independent server services.
func Sub_4EC720() {
	outer := GetServer()
	srv := outer.S()
	respawnSchedulerNative4EC720(respawnSchedulerRuntime4EC720{
		LookupTypeByName: func(name string) uint32 {
			return uint32(srv.Types.IndByID(name))
		},
		GameFlagsCheck: func(mask uint32) uint32 {
			return uint32(bool2int(noxflags.HasGame(noxflags.GameFlag(mask))))
		},
		UnitDefAllowed: func(typeInd uint32) bool {
			return srv.Types.ByInd(int(typeInd)).Allowed()
		},
		NewObjectByTypeInd: func(typeInd uint32) *server.Object {
			return srv.Objs.NewObject(srv.Types.ByInd(int(typeInd)))
		},
		SpecialModeCheck: func(mask uint32) uint32 {
			if mask == 0x2000 && noxflags.HasGame(1056) {
				return 1
			}
			return uint32(bool2int(mask&Nox_xxx_getServerSubFlags_409E60() != 0))
		},
		FPS:   gameFPSHook,
		Frame: gameFrameHook,
		PointFX: func(code uint32, pos types.Pointf) {
			srv.Nox_xxx_netSendPointFx_522FF0(netmsg.Op(code), pos)
		},
		AudioAtPosition: func(code uint32, pos types.Pointf) {
			srv.Audio.EventPos(sound.ID(code), pos, 0, 0)
		},
		AudioOnObject: func(code uint32, obj *server.Object) {
			srv.Audio.EventObj(sound.ID(code), obj, 0, 0)
		},
		MoveTo:           Nox_xxx_unitMove_4E7010,
		WeaponEquipFlags: objectNPCWeaponEquipFlags,
		SetHP:            Nox_xxx_unitSetHP_4E4560,
		CreateAt: func(obj, owner *server.Object, pos types.Pointf) {
			outer.CreateObjectAt(obj, owner, pos)
		},
		CopyModifierAttrs: Nox_xxx_modifSetItemAttrs_4E4990,
		DelayedDelete:     outer.DelayedDelete,
	})
}

//export sub_4EC720
func sub_4EC720() {
	Sub_4EC720()
}

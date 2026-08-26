//go:build amd64 || arm64

package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"fmt"

	"github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const npcXferVersion52ADE0 = 62

//export nox_xxx_XFerNPC_native_52ADE0
func nox_xxx_XFerNPC_native_52ADE0(objC *nox_object_t) int32 {
	obj := asObjectS(objC)
	if err := xferNPCNative52ADE0(cryptfile.Global(), obj); err != nil {
		mapLog.Printf("nox_xxx_XFerNPC_52ADE0: %v", err)
		return 0
	}
	return 1
}

// xferNPCNative52ADE0 preserves the GAME.EXE byte stream while keeping every
// live Object, MonsterUpdateData and inventory link at the host pointer width.
// The original PE32 implementation remains active on 32-bit builds.
func xferNPCNative52ADE0(cf *cryptfile.CryptFile, obj *server.Object) error {
	if cf == nil || obj == nil || obj.UpdateData == nil {
		return fmt.Errorf("missing crypt file, object, or NPC update data")
	}
	ud := obj.UpdateDataMonster()
	savedField34 := obj.Field34

	version, err := monsterRWU16(cf, npcXferVersion52ADE0)
	if err != nil {
		return fmt.Errorf("version: %w", err)
	}
	if version > npcXferVersion52ADE0 {
		return fmt.Errorf("unsupported version %d", version)
	}
	// Contemporary Nox maps use version 62. Older formats have several
	// different fixed-width branches; the untouched PE32 implementation keeps
	// supporting those on 32-bit targets until each old stream is recovered.
	if version < 50 {
		return fmt.Errorf("NPC transfer version %d still requires the PE32 path", version)
	}

	if cf.ReadOnly() {
		ud.MonsterDef = Nox_xxx_monsterDefByTT_517560(int(obj.TypeInd))
		if ud.MonsterDef != nil {
			ud.StatusFlags = ud.MonsterDef.StatusFlags92
		}
	}
	if Nox_xxx_mapReadWriteObjData_4F4530(obj, int(version)) == 0 {
		return fmt.Errorf("base object transfer failed")
	}
	if err := monsterXferDirection528DB0(cf, obj); err != nil {
		return err
	}
	if err := monsterXferCallbacks528DB0(cf, obj); err != nil {
		return err
	}
	if _, err := monsterRWU32(cf, 0); err != nil {
		return fmt.Errorf("direction scratch: %w", err)
	}
	if err := npcXferColors52ADE0(cf, obj); err != nil {
		return err
	}
	if err := monsterXferDefinitionCommon528DB0(cf, ud, obj.HealthData, int(version), true); err != nil {
		return fmt.Errorf("NPC definition: %w", err)
	}

	refs := new(monsterXferRefs528DB0)
	if err := monsterXferActionData529CE0(cf, obj, refs); err != nil {
		return err
	}
	if err := npcXferTail52ADE0(cf, obj, int(version)); err != nil {
		return err
	}
	if obj.Field34 != 0 && cf.ReadOnly() {
		if err := monsterXferInventory4F3E30(cf, obj, int(version), obj.Field34); err != nil {
			return err
		}
	}

	npcNormalizeEquipped52BA70(obj)
	if cf.ReadOnly() && noxflags.HasGame(noxflags.GameHost) {
		npcRestoreEquipped52ADE0(obj)
	}
	if cf.ReadOnly() {
		monsterXferPending528DB0.Lock()
		monsterXferPending528DB0.m[obj] = refs
		monsterXferPending528DB0.Unlock()
	}
	obj.Field34 = savedField34
	return nil
}

func npcXferColors52ADE0(cf *cryptfile.CryptFile, obj *server.Object) error {
	ud := obj.UpdateDataMonster()
	for i := range ud.Color {
		cl := ud.Color[i]
		buf := []byte{cl.R, cl.G, cl.B}
		if err := monsterRWBytes528DB0(cf, buf); err != nil {
			return fmt.Errorf("NPC color[%d]: %w", i, err)
		}
		if cf.ReadOnly() {
			cl = server.Color3{R: buf[0], G: buf[1], B: buf[2]}
			obj.Nox_xxx_setNPCColor_4E4A90(byte(i), &cl)
		}
	}
	return nil
}

func npcXferTail52ADE0(cf *cryptfile.CryptFile, obj *server.Object, version int) error {
	ud := obj.UpdateDataMonster()
	if obj.HealthData == nil {
		return fmt.Errorf("NPC has no health data")
	}
	var err error

	field361High := byte(ud.Field361 >> 8)
	if field361High, err = monsterRWU8(cf, field361High); err != nil {
		return fmt.Errorf("field361 high byte: %w", err)
	}
	ud.Field361 = ud.Field361&0xffff00ff | uint32(field361High)<<8
	if ud.Field0, err = monsterRWU32(cf, ud.Field0); err != nil {
		return fmt.Errorf("field0: %w", err)
	}
	healthMax, err := monsterRWU32(cf, uint32(obj.HealthData.Max))
	if err != nil {
		return fmt.Errorf("maximum health: %w", err)
	}
	obj.HealthData.Max = uint16(healthMax)
	subclassFlags, err := monsterRWU32(cf, uint32(obj.ObjSubClass)&0x180)
	if err != nil {
		return fmt.Errorf("NPC subclass flags: %w", err)
	}
	obj.ObjSubClass |= object.SubClass(subclassFlags)
	if obj.HealthData.Cur, err = monsterRWU16(cf, obj.HealthData.Cur); err != nil {
		return fmt.Errorf("health: %w", err)
	}
	if obj.Experience, err = monsterRWF32(cf, obj.Experience); err != nil {
		return fmt.Errorf("experience: %w", err)
	}
	if err := monsterXferVoice528DB0(cf, ud); err != nil {
		return fmt.Errorf("NPC voice: %w", err)
	}
	if err := monsterXferBuffs52AAB0(cf, obj); err != nil {
		return err
	}
	poison, err := monsterRWU8(cf, obj.Poison540)
	if err != nil {
		return fmt.Errorf("poison: %w", err)
	}
	if !cf.ReadOnly() {
		return nil
	}
	if poison != 0 {
		setSomePoisonDataCall4EEA90(obj, int32(poison))
	}
	if field361High != 0 && noxflags.HasGame(noxflags.GameHost) {
		obj.HealthData.Cur = 0
		obj.HealthData.Field2 = 0
	}
	if uint32(obj.ObjFlags)&0x8000 != 0 {
		obj.ObjFlags |= object.Flags(0x40)
	}
	_ = version
	return nil
}

// npcNormalizeEquipped52BA70 keeps only the first mutually exclusive
// two-handed weapon or shield, matching GAME.EXE 0052BA70.
func npcNormalizeEquipped52BA70(obj *server.Object) {
	if obj == nil {
		return
	}
	weaponSeen := false
	shieldSeen := false
	for item := obj.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.ObjFlags.Has(object.FlagEquipped) {
			continue
		}
		if item.ObjClass.HasAny(object.ClassWeapon|object.ClassWand) && uint32(item.ObjSubClass)&0x07ffe40c != 0 {
			if shieldSeen {
				item.ObjFlags &^= object.FlagEquipped
			} else {
				weaponSeen = true
			}
		} else if item.ObjClass.Has(object.ClassArmor) && uint8(item.ObjSubClass)&2 != 0 {
			if weaponSeen {
				item.ObjFlags &^= object.FlagEquipped
			} else {
				shieldSeen = true
			}
		}
	}
}

// npcRestoreEquipped52ADE0 restores the synchronized NPC appearance without
// placing a native pointer in the original PE32 Field516 slot. The inventory
// links and equipped flags remain native-width state.
func npcRestoreEquipped52ADE0(obj *server.Object) {
	if obj == nil || obj.UpdateData == nil {
		return
	}
	ud := obj.UpdateDataMonster()
	ud.WeaponEquipFlags = 0
	ud.ArmorEquipFlags = 0
	for item := obj.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.ObjFlags.Has(object.FlagEquipped) {
			continue
		}
		if uint8(obj.ObjSubClass)&0x10 != 0 {
			obj.SetNPCItemEquipFlags(item, true, objectNPCWeaponEquipFlags, objectNPCArmorEquipFlags)
		}
	}
}

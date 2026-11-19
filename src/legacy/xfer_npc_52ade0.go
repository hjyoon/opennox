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

//export nox_xxx_XFerNPC_52ADE0
func nox_xxx_XFerNPC_52ADE0(objC *nox_object_t) C.int {
	obj := asObjectS(objC)
	if err := xferNPCNative52ADE0(cryptfile.Global(), obj); err != nil {
		mapLog.Printf("nox_xxx_XFerNPC_52ADE0: %v", err)
		return 0
	}
	return 1
}

// xferNPCNative52ADE0 preserves the GAME.EXE byte stream while keeping every
// live Object, MonsterUpdateData and inventory link at the host pointer width.
func xferNPCNative52ADE0(cf *cryptfile.CryptFile, obj *server.Object) error {
	if cf == nil || obj == nil || obj.UpdateData == nil {
		return fmt.Errorf("missing crypt file, object, or NPC update data")
	}
	if cf.ReadOnly() {
		obj.TakeMonsterActionRefs()
	}
	ud := obj.UpdateDataMonster()
	savedField34 := obj.Field34

	versionWord, err := monsterRWU16(cf, npcXferVersion52ADE0)
	if err != nil {
		return fmt.Errorf("version: %w", err)
	}
	version := int(int16(versionWord))
	if version > npcXferVersion52ADE0 {
		return fmt.Errorf("unsupported version %d", version)
	}

	if cf.ReadOnly() {
		ud.MonsterDef = Nox_xxx_monsterDefByTT_517560(int(obj.TypeInd))
		if ud.MonsterDef != nil {
			ud.StatusFlags = ud.MonsterDef.StatusFlags92
		}
	}
	if objectMapReadWriteNative4F4530(cf, obj, int32(version)) == 0 {
		return fmt.Errorf("base object transfer failed")
	}
	if err := monsterXferDirectionAndCallbacks528DB0(cf, obj, version, true); err != nil {
		return err
	}
	if err := npcXferColors52ADE0(cf, obj); err != nil {
		return err
	}
	if cf.ReadOnly() && version == 31 {
		if err := npcXferVersion31Spells52ADE0(cf); err != nil {
			return err
		}
	}
	if err := monsterXferDefinitionCommon528DB0(cf, ud, obj.HealthData, version, true); err != nil {
		return fmt.Errorf("NPC definition: %w", err)
	}

	refs := new(server.MonsterActionRefs)
	if version >= 41 {
		if err := monsterXferActionData529CE0(cf, obj, refs, false); err != nil {
			return err
		}
	}
	if err := npcXferTail52ADE0(cf, obj, version); err != nil {
		return err
	}
	if obj.Field34 != 0 && cf.ReadOnly() {
		if err := monsterXferInventory4F3E30(cf, obj, version, obj.Field34); err != nil {
			return err
		}
	}

	npcNormalizeEquipped52BA70(obj)
	if cf.ReadOnly() && noxflags.HasGame(noxflags.GameHost) {
		npcRestoreEquipped52ADE0(obj)
	}
	if cf.ReadOnly() {
		obj.StoreMonsterActionRefs(refs)
	}
	obj.Field34 = savedField34
	return nil
}

func npcXferVersion31Spells52ADE0(cf *cryptfile.CryptFile) error {
	if _, err := monsterRWU16(cf, 0); err != nil {
		return fmt.Errorf("version 31 NPC spell prefix: %w", err)
	}
	count, err := monsterRWU8(cf, 0)
	if err != nil {
		return fmt.Errorf("version 31 NPC spell count: %w", err)
	}
	for i := 0; i < int(count); i++ {
		if _, err := monsterRWString8(cf, ""); err != nil {
			return fmt.Errorf("version 31 NPC spell[%d] name: %w", i, err)
		}
		for j := 0; j < 2; j++ {
			if _, err := monsterRWU8(cf, 0); err != nil {
				return fmt.Errorf("version 31 NPC spell[%d] byte %d: %w", i, j, err)
			}
		}
	}
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
	var err error

	field361High := byte(ud.Field361 >> 8)
	if version >= 42 {
		if field361High, err = monsterRWU8(cf, field361High); err != nil {
			return fmt.Errorf("field361 high byte: %w", err)
		}
		ud.Field361 = ud.Field361&0xffff00ff | uint32(field361High)<<8
	}
	if version >= 44 {
		if ud.Field0, err = monsterRWU32(cf, ud.Field0); err != nil {
			return fmt.Errorf("field0: %w", err)
		}
	}
	health := obj.HealthData
	if version >= 45 {
		var maximum uint32
		if health != nil {
			maximum = uint32(health.Max)
		}
		maximum, err = monsterRWU32(cf, maximum)
		if err != nil {
			return fmt.Errorf("maximum health: %w", err)
		}
		if health != nil {
			health.Max = uint16(maximum)
		}
	}
	if version >= 46 {
		subclassFlags, err := monsterRWU32(cf, uint32(obj.ObjSubClass)&0x180)
		if err != nil {
			return fmt.Errorf("NPC subclass flags: %w", err)
		}
		obj.ObjSubClass |= object.SubClass(subclassFlags)
	}
	if version >= 48 {
		if health == nil {
			return fmt.Errorf("NPC has no health data")
		}
		if health.Cur, err = monsterRWU16(cf, health.Cur); err != nil {
			return fmt.Errorf("health: %w", err)
		}
	}
	// GAME.EXE 0052B8DA..0052B965 gates each later field separately.
	// In particular, version-60 Kalen ends after voice, before buffs/poison.
	if version >= 51 {
		if obj.Experience, err = monsterRWF32(cf, obj.Experience); err != nil {
			return fmt.Errorf("experience: %w", err)
		}
	}
	if version >= 52 {
		if err := monsterXferVoice528DB0(cf, ud); err != nil {
			return fmt.Errorf("NPC voice: %w", err)
		}
	}
	if version >= 61 {
		if err := monsterXferBuffs52AAB0(cf, obj); err != nil {
			return err
		}
	}
	var poison byte
	if version >= 62 {
		if poison, err = monsterRWU8(cf, obj.Poison540); err != nil {
			return fmt.Errorf("poison: %w", err)
		}
	}
	if !cf.ReadOnly() {
		return nil
	}
	if poison != 0 {
		setSomePoisonDataCall4EEA90(obj, int32(poison))
	}
	if field361High != 0 && noxflags.HasGame(noxflags.GameHost) {
		npcApplyDormantHealth52ADE0(obj.HealthData)
	}
	if uint32(obj.ObjFlags)&0x8000 != 0 {
		obj.ObjFlags |= object.Flags(0x40)
	}
	return nil
}

// npcApplyDormantHealth52ADE0 preserves GAME.EXE 0052ADE0's two word stores:
// PE32 health offsets 0 and 4 are Cur and Max. Field2 at offset 2 is not
// touched. Keeping this separate makes an otherwise easy Cur/Field2/Max
// transcription error directly testable on every native architecture.
func npcApplyDormantHealth52ADE0(health *server.HealthData) {
	health.Cur = 0
	health.Max = 0
}

// npcRestoreEquipped52ADE0 replays the original equip operations, restoring
// clothing's NoAutoDrop flag, armor value, engage effects, and appearance.
// The native equip helpers never narrow inventory or update-data pointers.
func npcRestoreEquipped52ADE0(obj *server.Object) {
	if obj == nil || obj.UpdateData == nil {
		return
	}
	for item := obj.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.ObjFlags.Has(object.FlagEquipped) {
			continue
		}
		class := item.ObjClass
		item.ObjFlags &^= object.FlagEquipped
		if class.HasAny(object.ClassWeapon | object.ClassWand) {
			npcWeaponEquipNative53A2C0(obj, item)
		} else {
			npcArmorEquipNative53E520(obj, item)
		}
	}
}

//go:build amd64 || arm64

package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
#include "server__script__script.h"
*/
import "C"

import (
	"fmt"
	"math/bits"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const monsterXferVersion528DB0 = 64

func monsterParseSpellID528DB0(name string, allowEmpty bool) (spell.ID, error) {
	if name == "" && allowEmpty {
		return spell.SPELL_INVALID, nil
	}
	id := spell.ParseID(name)
	// ParseID returns zero both for an unknown name and for the serialized
	// SPELL_INVALID sentinel. The latter is present in original Nox map data
	// and must remain a valid zero-valued spell reference.
	if id == spell.SPELL_INVALID && name != spell.SPELL_INVALID.String() {
		return spell.SPELL_INVALID, fmt.Errorf("unknown spell %q", name)
	}
	return id, nil
}

func monsterParseAbilityID528DB0(name string) (server.Ability, error) {
	for id, candidate := range server.AbilityNames {
		if candidate == name {
			return server.Ability(id), nil
		}
	}
	return server.AbilityInvalid, fmt.Errorf("unknown ability %q", name)
}

// Shop parameter strings are polymorphic: FieldGuide stores an object type,
// AbilityReward stores an ability, and every other shop item stores a spell.
func monsterParseShopParam528DB0(
	typ *server.ObjectType,
	name string,
	objectIndex func(string) (uint32, bool),
) (uint32, error) {
	if name == "" {
		return 0, nil
	}
	switch typ.XferFunc() {
	case Get_nox_xxx_XFerFieldGuide_4F6390():
		id, ok := objectIndex(name)
		if !ok {
			return 0, fmt.Errorf("unknown field-guide creature %q", name)
		}
		return id, nil
	case Get_nox_xxx_XFerAbilityReward_4F6240():
		id, err := monsterParseAbilityID528DB0(name)
		return uint32(id), err
	default:
		id, err := monsterParseSpellID528DB0(name, false)
		return uint32(id), err
	}
}

func monsterShopParamName528DB0(
	typ *server.ObjectType,
	param uint32,
	objectName func(uint32) (string, bool),
) (string, error) {
	if param == 0 {
		return "", nil
	}
	switch typ.XferFunc() {
	case Get_nox_xxx_XFerFieldGuide_4F6390():
		name, ok := objectName(param)
		if !ok {
			return "", fmt.Errorf("unknown field-guide creature type index %d", param)
		}
		return name, nil
	case Get_nox_xxx_XFerAbilityReward_4F6240():
		id := server.Ability(param)
		if id < server.AbilityInvalid || id >= server.AbilityMax {
			return "", fmt.Errorf("unknown ability ID %d", param)
		}
		return id.String(), nil
	default:
		id := spell.ID(param)
		name := id.String()
		if parsed, err := monsterParseSpellID528DB0(name, false); err != nil || parsed != id {
			return "", fmt.Errorf("unknown spell ID %d", param)
		}
		return name, nil
	}
}

func monsterParseShopModifierSlot528DB0(name string, modifierID func(string) int) uint32 {
	if name == "" || modifierID == nil {
		return 0
	}
	id := modifierID(name)
	slot, ok := server.EncodeShopkeeperModifierID(id)
	if !ok {
		return 0
	}
	return slot
}

func monsterShopModifierName528DB0(slot uint32, modifierName func(int) (string, bool)) (string, error) {
	id, present, valid := server.DecodeShopkeeperModifierID(slot)
	if !valid {
		return "", fmt.Errorf("invalid modifier token %#x", slot)
	}
	if !present {
		return "", nil
	}
	if modifierName == nil {
		return "", fmt.Errorf("modifier ID %d cannot be resolved", id)
	}
	name, ok := modifierName(id)
	if !ok || name == "" {
		return "", fmt.Errorf("unknown modifier ID %d", id)
	}
	return name, nil
}

//export nox_xxx_XFerMonster_native_528DB0
func nox_xxx_XFerMonster_native_528DB0(objC *nox_object_t) int32 {
	obj := asObjectS(objC)
	if err := xferMonsterNative528DB0(cryptfile.Global(), obj); err != nil {
		mapLog.Printf("nox_xxx_XFerMonster_528DB0: %v", err)
		return 0
	}
	return 1
}

func xferMonsterNative528DB0(cf *cryptfile.CryptFile, obj *server.Object) error {
	if cf == nil || obj == nil || obj.UpdateData == nil {
		return fmt.Errorf("missing crypt file, object, or monster update data")
	}
	if cf.ReadOnly() {
		obj.TakeMonsterActionRefs()
	}
	ud := obj.UpdateDataMonster()
	refs := new(server.MonsterActionRefs)
	savedField34 := obj.Field34

	version, err := monsterRWU16(cf, monsterXferVersion528DB0)
	if err != nil {
		return fmt.Errorf("version: %w", err)
	}
	if version > monsterXferVersion528DB0 {
		return fmt.Errorf("unsupported version %d", version)
	}
	// The native object base reader is pointer-width safe from object version
	// 61 onward. Contemporary maps use monster transfer version 64; retaining
	// the original C implementation on 32-bit targets preserves older maps.
	if version < 41 {
		return fmt.Errorf("monster transfer version %d still requires the PE32 path", version)
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
	if _, err := monsterRWU32(cf, 0); err != nil { // legacy direction scratch value
		return fmt.Errorf("direction scratch: %w", err)
	}
	if err := monsterXferDefinition528DB0(cf, ud, int(version)); err != nil {
		return err
	}
	if err := monsterXferActionData529CE0(cf, obj, refs, false); err != nil {
		return err
	}
	if err := monsterXferTail528DB0(cf, obj, int(version)); err != nil {
		return err
	}

	if obj.Field34 != 0 {
		if cf.ReadOnly() {
			if err := monsterXferInventory4F3E30(cf, obj, int(version), obj.Field34); err != nil {
				return err
			}
		}
	}
	if cf.ReadOnly() {
		if noxflags.HasGame(noxflags.GameFlag22) || !Nox_xxx_gameIsSwitchToSolo_4DB240() {
			monsterOnSpawnNative529BC0(obj, obj.Server)
		}
		monsterEnsureGlyph528DB0(obj)
	}
	if cf.ReadOnly() {
		obj.StoreMonsterActionRefs(refs)
	}
	obj.Field34 = savedField34
	return nil
}

func monsterXferDirection528DB0(cf *cryptfile.CryptFile, obj *server.Object) error {
	var dir [2]uint32
	if !cf.ReadOnly() {
		var indexed server.IndexedDirectionVector509E20
		server.IndexedDirection509E20(int32(obj.Direction1), &indexed)
		dir[0] = uint32(indexed.X)
		dir[1] = uint32(indexed.Y)
	}
	for i := range dir {
		v, err := monsterRWU32(cf, dir[i])
		if err != nil {
			return fmt.Errorf("direction[%d]: %w", i, err)
		}
		dir[i] = v
	}
	if cf.ReadOnly() {
		angle := server.Dir16(server.DirectionToAngle509E00(&server.DirectionInitData{
			X: int32(dir[0]),
			Y: int32(dir[1]),
		}))
		obj.Direction1 = angle
		obj.Direction2 = angle
	}
	return nil
}

func monsterXferCallbacks528DB0(cf *cryptfile.CryptFile, obj *server.Object) error {
	ud := obj.UpdateDataMonster()
	callbacks := []struct {
		name string
		cb   *server.ScriptCallback
		off  uintptr
	}{
		{"enemy sighted", &ud.ScriptEnemySighted, 640},
		{"death", &ud.ScriptDeath, 896},
	}
	for _, it := range callbacks {
		if err := monsterXferScriptCallback528DB0(obj, it.cb, it.off); err != nil {
			return fmt.Errorf("%s callback: %w", it.name, err)
		}
	}
	low, err := monsterRWU16(cf, uint16(ud.Field305))
	if err != nil {
		return fmt.Errorf("callback field305: %w", err)
	}
	ud.Field305 = ud.Field305&0xffff0000 | uint32(low)
	callbacks = []struct {
		name string
		cb   *server.ScriptCallback
		off  uintptr
	}{
		{"looking for enemy", &ud.ScriptLookingForEnemy, 768},
		{"change focus", &ud.ScriptChangeFocus, 1024},
		{"is hit", &ud.ScriptIsHit, 1152},
		{"retreat", &ud.ScriptRetreat, 1280},
		{"collision", &ud.ScriptCollision, 1408},
		{"hear enemy", &ud.ScriptHearEnemy, 1536},
		{"end waypoint", &ud.ScriptEndOfWaypoint, 1664},
		{"lost enemy", &ud.ScriptLostEnemy, 1792},
	}
	for _, it := range callbacks {
		if err := monsterXferScriptCallback528DB0(obj, it.cb, it.off); err != nil {
			return fmt.Errorf("%s callback: %w", it.name, err)
		}
	}
	return nil
}

func monsterXferScriptCallback528DB0(obj *server.Object, cb *server.ScriptCallback, contextOff uintptr) error {
	var context unsafe.Pointer
	if obj.Field189 != nil {
		context = unsafe.Add(obj.Field189, contextOff)
	}
	if scriptHandlerXferRuntime4F5580(cb, context) == 0 {
		return fmt.Errorf("script handler transfer failed")
	}
	return nil
}

func monsterXferDefinition528DB0(cf *cryptfile.CryptFile, ud *server.MonsterUpdateData, version int) error {
	return monsterXferDefinitionCommon528DB0(cf, ud, nil, version, false)
}

func monsterXferDefinitionCommon528DB0(cf *cryptfile.CryptFile, ud *server.MonsterUpdateData, health *server.HealthData, version int, npc bool) error {
	var err error
	if b, e := monsterRWU8(cf, byte(ud.Field333)); e != nil {
		return fmt.Errorf("field333: %w", e)
	} else {
		ud.Field333 = ud.Field333&0xffffff00 | uint32(b)
	}
	status, err := monsterRWU32(cf, uint32(ud.StatusFlags))
	if err != nil {
		return fmt.Errorf("status: %w", err)
	}
	ud.StatusFlags = object.MonsterStatus(status)
	for _, it := range []struct {
		name string
		p    *float32
	}{
		{"field338", &ud.Field338},
		{"retreat", &ud.RetreatLevel},
		{"resume", &ud.ResumeLevel},
		{"sight", &ud.SightRange},
	} {
		if *it.p, err = monsterRWF32(cf, *it.p); err != nil {
			return fmt.Errorf("%s: %w", it.name, err)
		}
	}
	if npc {
		if health == nil {
			return fmt.Errorf("NPC has no health data")
		}
		if health.Cur, err = monsterRWU16(cf, health.Cur); err != nil {
			return fmt.Errorf("NPC definition health: %w", err)
		}
	}
	if ud.Aggression, err = monsterRWF32(cf, ud.Aggression); err != nil {
		return fmt.Errorf("aggression: %w", err)
	}
	ud.Aggression2 = ud.Aggression

	buf := unsafe.Slice((*byte)(unsafe.Pointer(&ud.Field341)), 76)
	name, err := monsterRWString8(cf, cStringBytes528DB0(buf))
	if err != nil {
		return fmt.Errorf("definition name: %w", err)
	}
	if cf.ReadOnly() {
		setCStringBytes528DB0(buf, name)
	}

	spells := unsafe.Slice(&ud.Field372, server.SpellsMax)
	if cf.ReadOnly() {
		clear(spells)
	}
	count := uint32(0)
	if !cf.ReadOnly() {
		for _, value := range spells {
			if value != 0 {
				count++
			}
		}
	}
	if count, err = monsterRWU32(cf, count); err != nil {
		return fmt.Errorf("spell count: %w", err)
	}
	if count > server.SpellsMax {
		return fmt.Errorf("spell count %d exceeds %d", count, server.SpellsMax)
	}
	if cf.ReadOnly() {
		for i := uint32(0); i < count; i++ {
			spellName, e := monsterRWString8(cf, "")
			if e != nil {
				return fmt.Errorf("spell[%d] name: %w", i, e)
			}
			id, e := monsterParseSpellID528DB0(spellName, false)
			if e != nil {
				return fmt.Errorf("spell[%d]: %w", i, e)
			}
			if int(id) >= len(spells) {
				return fmt.Errorf("spell[%d] id %d exceeds %d", i, id, len(spells)-1)
			}
			value, e := monsterRWU32(cf, 0)
			if e != nil {
				return fmt.Errorf("spell[%d] value: %w", i, e)
			}
			spells[id] = value
		}
	} else {
		for id, value := range spells {
			if value == 0 {
				continue
			}
			if _, err = monsterRWString8(cf, spell.ID(id).String()); err != nil {
				return fmt.Errorf("spell[%d] name: %w", id, err)
			}
			if _, err = monsterRWU32(cf, value); err != nil {
				return fmt.Errorf("spell[%d] value: %w", id, err)
			}
		}
	}

	pairs := [][2]*uint16{
		{&ud.Field362_0, &ud.Field362_2},
		{&ud.Field364_0, &ud.Field364_2},
		{&ud.Field366_0, &ud.Field366_2},
		{&ud.Field368_0, &ud.Field368_2},
		{&ud.Field370_0, &ud.Field370_2},
	}
	for i := range pairs {
		for j := range pairs[i] {
			if *pairs[i][j], err = monsterRWU16(cf, *pairs[i][j]); err != nil {
				return fmt.Errorf("definition pair[%d][%d]: %w", i, j, err)
			}
		}
	}
	if ud.Field329, err = monsterRWF32(cf, ud.Field329); err != nil {
		return fmt.Errorf("field329: %w", err)
	}
	if ud.Field510, err = monsterRWU32(cf, ud.Field510); err != nil {
		return fmt.Errorf("field510: %w", err)
	}
	if npc {
		field331, e := monsterRWU8(cf, byte(ud.Field331))
		if e != nil {
			return fmt.Errorf("field331: %w", e)
		}
		ud.Field331 = ud.Field331&0xffffff00 | uint32(field331)
		if ud.Field332, err = monsterRWF32(cf, ud.Field332); err != nil {
			return fmt.Errorf("field332: %w", err)
		}
	}
	if ud.Field330, err = monsterRWF32(cf, ud.Field330); err != nil {
		return fmt.Errorf("field330: %w", err)
	}
	spellIDs := []*uint32{&ud.Field511, &ud.Field512, &ud.Field513}
	for i, p := range spellIDs {
		cur := ""
		if !cf.ReadOnly() {
			// The original name table includes SPELL_INVALID at index zero.
			cur = spell.ID(*p).String()
		}
		value, e := monsterRWString8(cf, cur)
		if e != nil {
			return fmt.Errorf("auto spell[%d]: %w", i, e)
		}
		if cf.ReadOnly() {
			id, e := monsterParseSpellID528DB0(value, true)
			if e != nil {
				return fmt.Errorf("auto spell[%d]: %w", i, e)
			}
			*p = uint32(id)
		}
	}
	actionName := monsterSelectedActionNameNative5345B0(ud.AIAction340)
	if cf.ReadOnly() {
		actionName = ""
	}
	actionName, err = monsterRWString8(cf, actionName)
	if err != nil {
		return fmt.Errorf("AI action: %w", err)
	}
	if cf.ReadOnly() {
		ud.AIAction340 = uint32(monsterActionByName528DB0(actionName, 39, ai.ACTION_INVALID))
	}
	_ = version
	return nil
}

func monsterXferTail528DB0(cf *cryptfile.CryptFile, obj *server.Object, version int) error {
	ud := obj.UpdateDataMonster()
	var err error
	field361High := byte(ud.Field361 >> 8)
	if field361High, err = monsterRWU8(cf, field361High); err != nil {
		return fmt.Errorf("field361 high byte: %w", err)
	}
	ud.Field361 = ud.Field361&0xffff00ff | uint32(field361High)<<8
	if obj.MonsterClass().Has(object.MonsterShopkeeper) {
		if err := monsterXferShop528DB0(cf, obj, version); err != nil {
			return err
		}
	}
	if ud.Field0, err = monsterRWU32(cf, ud.Field0); err != nil {
		return fmt.Errorf("field0: %w", err)
	}
	serializedFlags, err := monsterRWU32(cf, uint32(obj.ObjSubClass)&0x180)
	if err != nil {
		return fmt.Errorf("monster subclass flags: %w", err)
	}
	obj.ObjSubClass |= object.SubClass(serializedFlags)
	if obj.HealthData == nil {
		return fmt.Errorf("monster has no health data")
	}
	if obj.HealthData.Cur, err = monsterRWU16(cf, obj.HealthData.Cur); err != nil {
		return fmt.Errorf("health: %w", err)
	}
	for _, it := range []struct {
		name string
		p    *uint32
	}{
		{"field337", &ud.Field337},
		{"field335", &ud.Field335},
		{"field361", &ud.Field361},
		{"field509", &ud.Field509},
	} {
		b, e := monsterRWU8(cf, byte(*it.p))
		if e != nil {
			return fmt.Errorf("%s low byte: %w", it.name, e)
		}
		*it.p = *it.p&0xffffff00 | uint32(b)
	}
	if obj.MonsterClass().Has(object.MonsterFemaleNPC) {
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
		if err := monsterXferVoice528DB0(cf, ud); err != nil {
			return fmt.Errorf("female NPC voice: %w", err)
		}
	}
	// These fields were appended in different record versions. Reading them
	// from a version-60 map consumes the next object's aligned header.
	if version >= 62 {
		if err := monsterXferBuffs52AAB0(cf, obj); err != nil {
			return err
		}
	}
	if version >= 63 && obj.MonsterClass().Has(object.MonsterWoundedNPC) {
		if err := monsterXferVoice528DB0(cf, ud); err != nil {
			return fmt.Errorf("wounded NPC voice: %w", err)
		}
	}
	if version >= 64 {
		poison, err := monsterRWU8(cf, obj.Poison540)
		if err != nil {
			return fmt.Errorf("poison: %w", err)
		}
		if cf.ReadOnly() && poison != 0 {
			setSomePoisonDataCall4EEA90(obj, int32(poison))
		}
	}
	if !cf.ReadOnly() {
		return nil
	}
	if field361High != 0 && noxflags.HasGame(noxflags.GameHost) {
		// GAME.EXE 00529A3D and 00529A49 clear health offsets 4 and 0:
		// Max and Cur, not the previous-health word at offset 2. Leaving
		// Max nonzero makes invulnerable female NPCs retreat and eat food.
		obj.HealthData.Max = 0
		obj.HealthData.Cur = 0
	}
	if obj.ObjFlags&0x8000 != 0 && !obj.Server().IsZombie(obj) {
		obj.ObjFlags |= 0x40
	}
	return nil
}

func monsterXferShop528DB0(cf *cryptfile.CryptFile, obj *server.Object, version int) error {
	if obj.InitData == nil {
		return fmt.Errorf("shopkeeper has no init data")
	}
	shop := obj.InitDataShopkeeper()
	var err error
	if shop.BuyMultiplier, err = monsterRWF32(cf, shop.BuyMultiplier); err != nil {
		return fmt.Errorf("shop buy multiplier: %w", err)
	}
	if version >= 61 {
		if shop.SellMultiplier, err = monsterRWF32(cf, shop.SellMultiplier); err != nil {
			return fmt.Errorf("shop sell multiplier: %w", err)
		}
	}
	text, err := monsterRWString8(cf, cStringBytes528DB0(shop.ShopText[:]))
	if err != nil {
		return fmt.Errorf("shop text: %w", err)
	}
	if len(text) >= len(shop.ShopText) {
		return fmt.Errorf("shop text length %d exceeds %d", len(text), len(shop.ShopText)-1)
	}
	if cf.ReadOnly() {
		setCStringBytes528DB0(shop.ShopText[:], text)
	}
	if shop.Count, err = monsterRWU8(cf, shop.Count); err != nil {
		return fmt.Errorf("shop item count: %w", err)
	}
	if int(shop.Count) > len(shop.Items) {
		return fmt.Errorf("shop item count %d exceeds %d", shop.Count, len(shop.Items))
	}
	for i := 0; i < int(shop.Count); i++ {
		if err := monsterXferShopItem528DB0(cf, obj.Server(), &shop.Items[i], version); err != nil {
			return fmt.Errorf("shop item[%d]: %w", i, err)
		}
	}
	return nil
}

func monsterXferShopItem528DB0(cf *cryptfile.CryptFile, srv *server.Server, item *server.ShopkeeperItemDefinition, version int) error {
	var err error
	if version < 50 {
		if _, err = monsterRWU32(cf, 0); err != nil {
			return err
		}
	}
	if item.Count, err = monsterRWU8(cf, item.Count); err != nil {
		return err
	}
	typeName := ""
	var typ *server.ObjectType
	if !cf.ReadOnly() {
		typ = srv.Types.ByInd(int(item.TypeInd))
		if typ == nil {
			return fmt.Errorf("unknown object type index %d", item.TypeInd)
		}
		typeName = typ.ID()
	}
	typeName, err = monsterRWString8(cf, typeName)
	if err != nil {
		return err
	}
	if cf.ReadOnly() {
		typ = srv.Types.ByID(typeName)
		if typ == nil {
			return fmt.Errorf("unknown object type %q", typeName)
		}
		item.TypeInd = uint32(typ.Ind())
	}
	if version >= 47 {
		paramName := ""
		paramPresent := true
		if !cf.ReadOnly() {
			paramName, err = monsterShopParamName528DB0(typ, item.Param, func(index uint32) (string, bool) {
				paramType := srv.Types.ByInd(int(index))
				if paramType == nil {
					return "", false
				}
				return paramType.ID(), true
			})
			if err != nil {
				return fmt.Errorf("shop parameter: %w", err)
			}
			if item.Param == 0 {
				// A missing field-guide type is a NULL name pointer, but the
				// spell/ability tables have named invalid sentinels at zero.
				switch typ.XferFunc() {
				case Get_nox_xxx_XFerFieldGuide_4F6390():
					paramPresent = false
				case Get_nox_xxx_XFerAbilityReward_4F6240():
					paramName = server.AbilityInvalid.String()
				default:
					paramName = spell.SPELL_INVALID.String()
				}
			}
		}
		paramName, err = monsterRWOptionalString8(cf, paramName, paramPresent)
		if err != nil {
			return err
		}
		if cf.ReadOnly() {
			id, e := monsterParseShopParam528DB0(typ, paramName, func(name string) (uint32, bool) {
				paramType := srv.Types.ByID(name)
				if paramType == nil {
					return 0, false
				}
				return uint32(paramType.Ind()), true
			})
			if e != nil {
				return fmt.Errorf("shop parameter for %q: %w", typeName, e)
			}
			item.Param = id
		}
	}
	for i := range item.ModifierSlots {
		name := ""
		if !cf.ReadOnly() {
			name, err = monsterShopModifierName528DB0(item.ModifierSlots[i], func(id int) (string, bool) {
				modifier := srv.Modif.Nox_xxx_modifGetDescById413330(id)
				if modifier == nil {
					return "", false
				}
				return modifier.Name(), true
			})
			if err != nil {
				return fmt.Errorf("modifier slot %d: %w", i, err)
			}
		}
		name, err = monsterRWOptionalString8(cf, name, item.ModifierSlots[i] != 0)
		if err != nil {
			return err
		}
		if cf.ReadOnly() {
			item.ModifierSlots[i] = monsterParseShopModifierSlot528DB0(name, srv.Modif.Nox_xxx_modifGetIdByName413290)
		}
	}
	return nil
}

func monsterXferVoice528DB0(cf *cryptfile.CryptFile, ud *server.MonsterUpdateData) error {
	name := ""
	if !cf.ReadOnly() {
		name = soundSetName(ud.SoundSet122)
	}
	name, err := monsterRWOptionalString8(cf, name, ud.SoundSet122 != nil)
	if err != nil {
		return err
	}
	if cf.ReadOnly() {
		ud.SoundSet122 = getDefaultSoundSet(name)
	}
	return nil
}

func monsterXferBuffs52AAB0(cf *cryptfile.CryptFile, obj *server.Object) error {
	version, err := monsterRWU16(cf, 2)
	if err != nil {
		return fmt.Errorf("buff version: %w", err)
	}
	if version == 0 || version > 2 {
		return fmt.Errorf("unsupported buff version %d", version)
	}
	count := byte(bits.OnesCount32(obj.Buffs))
	if count, err = monsterRWU8(cf, count); err != nil {
		return fmt.Errorf("buff count: %w", err)
	}
	if count > 32 {
		return fmt.Errorf("buff count %d exceeds 32", count)
	}
	if cf.ReadOnly() {
		for i := byte(0); i < count; i++ {
			name, e := monsterRWString8(cf, "")
			if e != nil {
				return fmt.Errorf("buff[%d] name: %w", i, e)
			}
			enc, ok := server.ParseEnchant(name)
			if !ok {
				return fmt.Errorf("buff[%d] has unknown name %q", i, name)
			}
			power, e := monsterRWU8(cf, 0)
			if e != nil {
				return fmt.Errorf("buff[%d] power: %w", i, e)
			}
			duration, e := monsterRWU32(cf, 0)
			if e != nil {
				return fmt.Errorf("buff[%d] duration: %w", i, e)
			}
			arg := server.SpellAcceptArg{Obj: obj, Pos: obj.Pos()}
			GetServer().Nox_xxx_spellAccept4FD400(enc.Spell(), obj, obj, obj, &arg, int(power))
			obj.BuffsDur[enc] = uint16(duration)
			obj.BuffsPower[enc] = power
			if enc == server.ENCHANT_SHIELD && version >= 2 {
				extra, e := monsterRWU32(cf, 100)
				if e != nil {
					return fmt.Errorf("buff[%d] shield duration: %w", i, e)
				}
				if dur := monsterFindDurSpell528DB0(obj, 51); dur != nil {
					dur.Field72 = int32(extra)
				}
			}
		}
		return nil
	}
	for enc := server.EnchantID(0); enc < 32; enc++ {
		if !obj.HasEnchant(enc) {
			continue
		}
		if _, err = monsterRWString8(cf, enc.String()); err != nil {
			return err
		}
		if _, err = monsterRWU8(cf, byte(obj.EnchantPower(enc))); err != nil {
			return err
		}
		if _, err = monsterRWU32(cf, uint32(obj.EnchantDur(enc))); err != nil {
			return err
		}
		if enc == server.ENCHANT_SHIELD {
			extra := uint32(100)
			if dur := monsterFindDurSpell528DB0(obj, 51); dur != nil {
				extra = uint32(dur.Field72)
			}
			if _, err = monsterRWU32(cf, extra); err != nil {
				return err
			}
		}
	}
	return nil
}

func monsterFindDurSpell528DB0(obj *server.Object, spellID uint32) *server.DurSpell {
	for it := obj.Server().Spells.Dur.List; it != nil; it = it.Next {
		if it.Flags88&1 == 0 && it.Spell == spellID && it.Target48 == obj {
			return it
		}
	}
	return nil
}

func monsterEnsureGlyph528DB0(obj *server.Object) {
	if obj == nil || !obj.Class().Has(object.ClassMonster) || !obj.MonsterClass().Has(object.MonsterBomber) {
		return
	}
	ud := obj.UpdateDataMonster()
	glyphType := obj.Server().Types.GlyphID()
	for item := obj.InvFirstItem; item != nil; item = item.InvNextItem {
		if int(item.TypeInd) == glyphType {
			return
		}
	}
	spells := [...]uint32{ud.Field511, ud.Field512, ud.Field513}
	count := 0
	for _, id := range spells {
		if id != 0 {
			count++
		}
	}
	if count == 0 {
		return
	}
	glyph := obj.Server().NewObjectByTypeID("Glyph")
	if glyph == nil || glyph.InitData == nil {
		return
	}
	data := glyph.InitDataGlyph()
	for i := 0; i < count; i++ {
		data.Spells[i] = spells[i]
	}
	data.SpellsCnt = uint32(count)
	data.SpellArg = server.SpellAcceptArg{Pos: obj.Pos()}
	Nox_xxx_inventoryPutImpl_4F3070(obj, glyph, 1)
}

//export nox_server_resolveMonsterXferRefs_528DB0
func nox_server_resolveMonsterXferRefs_528DB0(objC *nox_object_t) {
	monsterResolveXferRefs528DB0(asObjectS(objC))
}

//go:build amd64 || arm64

package legacy

/*
#include "defs.h"
#include "GAME4_1.h"
#include "server__script__script.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/bits"
	"sync"
	"unsafe"

	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const (
	monsterXferVersion528DB0       = 64
	monsterActionXferVersion529CE0 = 4
)

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

var monsterXferPending528DB0 = struct {
	sync.Mutex
	m map[*server.Object]*monsterXferRefs528DB0
}{m: make(map[*server.Object]*monsterXferRefs528DB0)}

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
	ud := obj.UpdateDataMonster()
	refs := new(monsterXferRefs528DB0)
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
	if err := monsterXferActionData529CE0(cf, obj, refs); err != nil {
		return err
	}
	if err := monsterXferTail528DB0(cf, obj, int(version)); err != nil {
		return err
	}

	if cf.ReadOnly() {
		monsterXferPending528DB0.Lock()
		monsterXferPending528DB0.m[obj] = refs
		monsterXferPending528DB0.Unlock()
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
			monsterOnSpawnNative529BC0(obj)
		}
		monsterEnsureGlyph528DB0(obj)
	}
	obj.Field34 = savedField34
	return nil
}

func monsterXferDirection528DB0(cf *cryptfile.CryptFile, obj *server.Object) error {
	var dir [2]uint32
	if !cf.ReadOnly() {
		C.nox_xxx_xferIndexedDirection_509E20(C.int(obj.Direction1), (*C.int2)(unsafe.Pointer(&dir[0])))
	}
	for i := range dir {
		v, err := monsterRWU32(cf, dir[i])
		if err != nil {
			return fmt.Errorf("direction[%d]: %w", i, err)
		}
		dir[i] = v
	}
	if cf.ReadOnly() {
		angle := server.Dir16(C.nox_xxx_xferDirectionToAngle_509E00((*C.uint32_t)(unsafe.Pointer(&dir[0]))))
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
	var context *C.char
	if obj.Field189 != nil {
		context = (*C.char)(unsafe.Add(obj.Field189, contextOff))
	}
	if C.nox_xxx_xferReadScriptHandler_4F5580(unsafe.Pointer(cb), context) == 0 {
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
		if !cf.ReadOnly() && *p != 0 {
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
	actionName := ai.ActionType(ud.AIAction340).String()
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

func monsterXferActionData529CE0(cf *cryptfile.CryptFile, obj *server.Object, refs *monsterXferRefs528DB0) error {
	ud := obj.UpdateDataMonster()
	version, err := monsterRWU16(cf, monsterActionXferVersion529CE0)
	if err != nil {
		return fmt.Errorf("action version: %w", err)
	}
	if version > monsterActionXferVersion529CE0 {
		return fmt.Errorf("unsupported action version %d", version)
	}
	enabled := byte(0)
	if noxflags.HasGame(noxflags.GameHost) && !noxflags.HasGame(noxflags.GameFlag23) {
		enabled = 1
	}
	if version >= 2 {
		enabled, err = monsterRWU8(cf, enabled)
		if err != nil {
			return fmt.Errorf("action enabled: %w", err)
		}
		if enabled == 0 && version != 1 {
			return nil
		}
	}

	frame := obj.Server().Frame()
	storedFrame, err := monsterRWU32(cf, frame)
	if err != nil {
		return fmt.Errorf("action frame: %w", err)
	}
	delta := int32(frame - storedFrame)
	if _, err = monsterRWU32(cf, 0); err != nil {
		return fmt.Errorf("action scratch: %w", err)
	}
	if ud.Field2, err = monsterRWU32(cf, ud.Field2); err != nil {
		return fmt.Errorf("path count: %w", err)
	}
	if ud.Field2 > uint32(len(ud.Path)) {
		return fmt.Errorf("path count %d exceeds %d", ud.Field2, len(ud.Path))
	}
	for i := 0; i < int(ud.Field2); i++ {
		if ud.Path[i].X, err = monsterRWF32(cf, ud.Path[i].X); err != nil {
			return fmt.Errorf("path[%d].x: %w", i, err)
		}
		if ud.Path[i].Y, err = monsterRWF32(cf, ud.Path[i].Y); err != nil {
			return fmt.Errorf("path[%d].y: %w", i, err)
		}
	}
	if ud.Field67, err = monsterRWU32(cf, ud.Field67); err != nil {
		return fmt.Errorf("field67: %w", err)
	}
	if ud.Field68.X, err = monsterRWF32(cf, ud.Field68.X); err != nil {
		return fmt.Errorf("field68.x: %w", err)
	}
	if ud.Field68.Y, err = monsterRWF32(cf, ud.Field68.Y); err != nil {
		return fmt.Errorf("field68.y: %w", err)
	}
	if ud.Field70, err = monsterRWU32(cf, ud.Field70); err != nil {
		return fmt.Errorf("field70: %w", err)
	}
	field71, err := monsterRWU8(cf, byte(ud.Field71))
	if err != nil {
		return fmt.Errorf("field71: %w", err)
	}
	ud.Field71 = ud.Field71&0xffffff00 | uint32(field71)
	if cf.ReadOnly() {
		ud.Field70 = monsterShiftFrame528DB0(ud.Field70, delta)
	}

	if ud.Field74, err = monsterRWU32(cf, ud.Field74); err != nil {
		return fmt.Errorf("waypoint count: %w", err)
	}
	if ud.Field74 > uint32(len(ud.Waypoints)) {
		return fmt.Errorf("waypoint count %d exceeds %d", ud.Field74, len(ud.Waypoints))
	}
	for i := 0; i < int(ud.Field74); i++ {
		ind := uint32(0)
		if !cf.ReadOnly() && ud.Waypoints[i] != nil {
			ind = ud.Waypoints[i].Index
		}
		ind, err = monsterRWU32(cf, ind)
		if err != nil {
			return fmt.Errorf("waypoint[%d]: %w", i, err)
		}
		if cf.ReadOnly() {
			ud.Waypoints[i] = obj.Server().WPs.PendingByInd(int(ind))
		}
	}
	for _, it := range []struct {
		name string
		p    *uint32
	}{
		{"field91", &ud.Field91},
		{"field92", &ud.Field92},
		{"field93", &ud.Field93},
		{"direction94", &ud.Direction94},
	} {
		if *it.p, err = monsterRWU32(cf, *it.p); err != nil {
			return fmt.Errorf("%s: %w", it.name, err)
		}
	}
	if ud.Pos95.X, err = monsterRWF32(cf, ud.Pos95.X); err != nil {
		return fmt.Errorf("pos95.x: %w", err)
	}
	if ud.Pos95.Y, err = monsterRWF32(cf, ud.Pos95.Y); err != nil {
		return fmt.Errorf("pos95.y: %w", err)
	}
	soundName := sound.ID(ud.Field97).String()
	if cf.ReadOnly() {
		soundName = ""
	}
	soundName, err = monsterRWString8(cf, soundName)
	if err != nil {
		return fmt.Errorf("sound: %w", err)
	}
	if cf.ReadOnly() {
		ud.Field97 = uint32(sound.ByName(soundName))
	}
	if ud.Field99X, err = monsterRWF32(cf, ud.Field99X); err != nil {
		return fmt.Errorf("field99.x: %w", err)
	}
	if ud.Field99Y, err = monsterRWF32(cf, ud.Field99Y); err != nil {
		return fmt.Errorf("field99.y: %w", err)
	}
	if ud.Field101, err = monsterRWU32(cf, ud.Field101); err != nil {
		return fmt.Errorf("field101: %w", err)
	}
	if cf.ReadOnly() {
		ud.Field101 = monsterShiftFrame528DB0(ud.Field101, delta)
	}
	for _, it := range []struct {
		name string
		p    *uint8
	}{
		{"field120.1", &ud.Field120_1},
		{"field120.2", &ud.Field120_2},
		{"field120.3", &ud.Field120_3},
	} {
		if *it.p, err = monsterRWU8(cf, *it.p); err != nil {
			return fmt.Errorf("%s: %w", it.name, err)
		}
	}
	for _, it := range []struct {
		name string
		p    *uint32
	}{
		{"field124", &ud.Field124},
		{"field125", &ud.Field125},
		{"field126", &ud.Field126},
	} {
		if *it.p, err = monsterRWU32(cf, *it.p); err != nil {
			return fmt.Errorf("%s: %w", it.name, err)
		}
	}
	if cf.ReadOnly() {
		ud.Field124 = monsterShiftFrame528DB0(ud.Field124, delta)
	}
	if _, err = monsterRWU32(cf, 0); err != nil {
		return fmt.Errorf("field133 scratch: %w", err)
	}
	if ud.Field134, err = monsterRWU32(cf, ud.Field134); err != nil {
		return fmt.Errorf("field134: %w", err)
	}
	if ud.Field135, err = monsterRWU32(cf, ud.Field135); err != nil {
		return fmt.Errorf("field135: %w", err)
	}
	if cf.ReadOnly() {
		ud.Field134 = monsterShiftFrame528DB0(ud.Field134, delta)
		ud.Field135 = monsterShiftFrame528DB0(ud.Field135, delta)
	}
	stackInd, err := monsterRWU8(cf, byte(ud.AIStackInd))
	if err != nil {
		return fmt.Errorf("AI stack index: %w", err)
	}
	ud.AIStackInd = int8(stackInd)
	if ud.AIStackInd >= int8(len(ud.AIStack)) {
		return fmt.Errorf("AI stack index %d exceeds %d", ud.AIStackInd, len(ud.AIStack)-1)
	}
	for i := 0; i <= int(ud.AIStackInd); i++ {
		if err := monsterXferAIStackItem52A440(cf, &ud.AIStack[i], &refs.aiObjects[i], &refs.aiWPs[i], delta); err != nil {
			return fmt.Errorf("AI stack[%d]: %w", i, err)
		}
	}

	if ud.Field282_1, err = monsterRWU8(cf, ud.Field282_1); err != nil {
		return fmt.Errorf("seen-enemy count: %w", err)
	}
	if int(ud.Field282_1) > len(ud.SeenEnemies) {
		return fmt.Errorf("seen-enemy count %d exceeds %d", ud.Field282_1, len(ud.SeenEnemies))
	}
	for i := 0; i < int(ud.Field282_1); i++ {
		id := int32(0)
		if !cf.ReadOnly() && ud.SeenEnemies[i] != nil {
			id = ud.SeenEnemies[i].ScriptIDVal
		}
		id, err = monsterRWI32(cf, id)
		if err != nil {
			return fmt.Errorf("seen enemy[%d]: %w", i, err)
		}
		if cf.ReadOnly() {
			refs.seen[i] = id
			ud.SeenEnemies[i] = nil
		}
	}
	current := int32(0)
	if !cf.ReadOnly() && ud.CurrentEnemy != nil {
		current = ud.CurrentEnemy.ScriptIDVal
	}
	if current, err = monsterRWI32(cf, current); err != nil {
		return fmt.Errorf("current enemy: %w", err)
	}
	if cf.ReadOnly() {
		refs.current = current
		ud.CurrentEnemy = nil
	}
	if ud.Field301, err = monsterRWU32(cf, ud.Field301); err != nil {
		return fmt.Errorf("field301: %w", err)
	}
	if cf.ReadOnly() {
		ud.Field301 = monsterShiftFrame528DB0(ud.Field301, delta)
	}
	if ud.DialogStartFunc, err = monsterRWI32(cf, ud.DialogStartFunc); err != nil {
		return fmt.Errorf("dialog start: %w", err)
	}
	if ud.DialogEndFunc, err = monsterRWI32(cf, ud.DialogEndFunc); err != nil {
		return fmt.Errorf("dialog end: %w", err)
	}
	if ud.DialogFlags, err = monsterRWU8(cf, ud.DialogFlags); err != nil {
		return fmt.Errorf("dialog flags: %w", err)
	}
	if ud.DialogResult, err = monsterRWU8(cf, ud.DialogResult); err != nil {
		return fmt.Errorf("dialog result: %w", err)
	}
	portrait, err := monsterRWString8(cf, cStringBytes528DB0(ud.DialogPortraitBuf[:]))
	if err != nil {
		return fmt.Errorf("dialog portrait: %w", err)
	}
	if len(portrait) >= len(ud.DialogPortraitBuf) {
		return fmt.Errorf("dialog portrait length %d exceeds %d", len(portrait), len(ud.DialogPortraitBuf)-1)
	}
	if cf.ReadOnly() {
		setCStringBytes528DB0(ud.DialogPortraitBuf[:], portrait)
	}
	if version < 4 {
		return nil
	}
	return monsterXferActionDataV4Tail529CE0(cf, obj, refs, delta)
}

func monsterXferAIStackItem52A440(cf *cryptfile.CryptFile, item *server.AIStackItem, objectIDs *[2]int32, waypointIDs *[2]uint32, delta int32) error {
	actionName := ai.ActionType(item.Action).String()
	if cf.ReadOnly() {
		actionName = ""
	}
	var err error
	actionName, err = monsterRWString8(cf, actionName)
	if err != nil {
		return fmt.Errorf("action name: %w", err)
	}
	if cf.ReadOnly() {
		item.Action = uint32(monsterActionByName528DB0(actionName, 72, ai.ACTION_IDLE))
		item.Args = [4]uintptr{}
	}
	action := int(item.Action)
	if action < 0 || action >= 72 {
		return fmt.Errorf("invalid action %d", action)
	}
	argCount := int(memmap.Uint8(0x587000, uintptr(255604+16*action)))
	storedCount, err := monsterRWU8(cf, byte(argCount))
	if err != nil {
		return fmt.Errorf("argument count: %w", err)
	}
	if cf.ReadOnly() && int(storedCount) != argCount {
		return fmt.Errorf("action %d argument count %d, want %d", action, storedCount, argCount)
	}
	if argCount > 2 {
		return fmt.Errorf("action %d argument count %d exceeds native slots", action, argCount)
	}
	for i := 0; i < argCount; i++ {
		slot := 2 * i
		kind := memmap.Uint32(0x587000, uintptr(255608+4*(i+4*action)))
		switch kind {
		case 0:
			for j := 0; j < 2; j++ {
				value, e := monsterRWU32(cf, uint32(item.Args[slot+j]))
				if e != nil {
					return fmt.Errorf("argument %d.%d: %w", i, j, e)
				}
				item.Args[slot+j] = uintptr(value)
			}
		case 1:
			id := int32(0)
			if !cf.ReadOnly() {
				if target := item.ArgObj(slot); target != nil {
					id = target.ScriptIDVal
				}
			}
			id, err = monsterRWI32(cf, id)
			if err != nil {
				return fmt.Errorf("object argument %d: %w", i, err)
			}
			if cf.ReadOnly() {
				objectIDs[i] = id
				item.Args[slot] = 0
			}
		case 2:
			ind := uint32(0)
			if !cf.ReadOnly() && item.Args[slot] != 0 {
				ind = (*server.Waypoint)(unsafe.Pointer(item.Args[slot])).Index
			}
			ind, err = monsterRWU32(cf, ind)
			if err != nil {
				return fmt.Errorf("waypoint argument %d: %w", i, err)
			}
			if cf.ReadOnly() {
				waypointIDs[i] = ind
				item.Args[slot] = 0
			}
		case 3, 4, 6:
			value, e := monsterRWU32(cf, uint32(item.Args[slot]))
			if e != nil {
				return fmt.Errorf("argument %d kind %d: %w", i, kind, e)
			}
			item.Args[slot] = uintptr(value)
		case 5:
			value, e := monsterRWU32(cf, uint32(item.Args[slot]))
			if e != nil {
				return fmt.Errorf("frame argument %d: %w", i, e)
			}
			if cf.ReadOnly() {
				value = monsterShiftFrame528DB0(value, delta)
			}
			item.Args[slot] = uintptr(value)
		case 7:
			value, e := monsterRWU8(cf, byte(item.Args[slot]))
			if e != nil {
				return fmt.Errorf("byte argument %d: %w", i, e)
			}
			item.Args[slot] = uintptr(value)
		default:
			return fmt.Errorf("action %d argument %d has unknown kind %d", action, i, kind)
		}
	}
	if item.Field5, err = monsterRWU32(cf, item.Field5); err != nil {
		return fmt.Errorf("field5: %w", err)
	}
	return nil
}

func monsterXferActionDataV4Tail529CE0(cf *cryptfile.CryptFile, obj *server.Object, refs *monsterXferRefs528DB0, delta int32) error {
	ud := obj.UpdateDataMonster()
	var err error
	for _, it := range []struct {
		name string
		p    *uint32
	}{
		{"field1", &ud.Field1},
		{"field72", &ud.Field72},
		{"field73", &ud.Field73},
	} {
		if *it.p, err = monsterRWU32(cf, *it.p); err != nil {
			return fmt.Errorf("%s: %w", it.name, err)
		}
	}
	field98ID := int32(0)
	if !cf.ReadOnly() {
		if target := monsterObjectByNetCode528DB0(obj.Server(), ud.Field98); target != nil {
			field98ID = target.ScriptIDVal
		}
	}
	if field98ID, err = monsterRWI32(cf, field98ID); err != nil {
		return fmt.Errorf("field98 reference: %w", err)
	}
	if cf.ReadOnly() {
		refs.field98 = field98ID
		ud.Field98 = 0
	}
	if ud.Field123, err = monsterRWU32(cf, ud.Field123); err != nil {
		return fmt.Errorf("field123: %w", err)
	}
	for _, it := range []struct {
		name  string
		p     *uint32
		frame bool
	}{
		{"field127", &ud.Field127, true},
		{"field128", &ud.Field128, true},
		{"field129", &ud.Field129, true},
		{"field130", &ud.Field130, true},
		{"field132", &ud.Field132, true},
		{"field133", &ud.Field133, true},
		{"field131", &ud.Field131, false},
		{"field137", &ud.Field137, true},
	} {
		if *it.p, err = monsterRWU32(cf, *it.p); err != nil {
			return fmt.Errorf("%s: %w", it.name, err)
		}
		if cf.ReadOnly() && it.frame {
			*it.p = monsterShiftFrame528DB0(*it.p, delta)
		}
	}
	if ud.Field282_0, err = monsterRWU8(cf, ud.Field282_0); err != nil {
		return fmt.Errorf("field282.0: %w", err)
	}
	field300ID := int32(0)
	if !cf.ReadOnly() {
		if target := monsterObjectByNetCode528DB0(obj.Server(), ud.Field300); target != nil {
			field300ID = target.ScriptIDVal
		}
	}
	if field300ID, err = monsterRWI32(cf, field300ID); err != nil {
		return fmt.Errorf("field300 reference: %w", err)
	}
	if cf.ReadOnly() {
		refs.field300 = field300ID
		ud.Field300 = 0
	}
	if ud.Field302, err = monsterRWU32(cf, ud.Field302); err != nil {
		return fmt.Errorf("field302: %w", err)
	}
	if ud.Field303, err = monsterRWU32(cf, ud.Field303); err != nil {
		return fmt.Errorf("field303: %w", err)
	}
	if cf.ReadOnly() {
		ud.Field302 = monsterShiftFrame528DB0(ud.Field302, delta)
		ud.Field303 = monsterShiftFrame528DB0(ud.Field303, delta)
	}
	preferred := int32(0)
	if !cf.ReadOnly() && ud.PreferredEnemy != nil {
		preferred = ud.PreferredEnemy.ScriptIDVal
	}
	if preferred, err = monsterRWI32(cf, preferred); err != nil {
		return fmt.Errorf("preferred enemy: %w", err)
	}
	if cf.ReadOnly() {
		refs.preferred = preferred
		ud.PreferredEnemy = nil
	}
	if ud.Field543_0, err = monsterRWU8(cf, ud.Field543_0); err != nil {
		return fmt.Errorf("field535 reference count: %w", err)
	}
	if int(ud.Field543_0) > len(ud.Field535) {
		return fmt.Errorf("field535 reference count %d exceeds %d", ud.Field543_0, len(ud.Field535))
	}
	for i := 0; i < int(ud.Field543_0); i++ {
		id := int32(0)
		if !cf.ReadOnly() {
			if target := monsterObjectByNetCode528DB0(obj.Server(), ud.Field535[i]); target != nil {
				id = target.ScriptIDVal
			}
		}
		id, err = monsterRWI32(cf, id)
		if err != nil {
			return fmt.Errorf("field535 reference[%d]: %w", i, err)
		}
		if cf.ReadOnly() {
			refs.field535[i] = id
			ud.Field535[i] = 0
		}
	}
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
	if err := monsterXferBuffs52AAB0(cf, obj); err != nil {
		return err
	}
	if obj.MonsterClass().Has(object.MonsterWoundedNPC) {
		if err := monsterXferVoice528DB0(cf, ud); err != nil {
			return fmt.Errorf("wounded NPC voice: %w", err)
		}
	}
	poison, err := monsterRWU8(cf, obj.Poison540)
	if err != nil {
		return fmt.Errorf("poison: %w", err)
	}
	if cf.ReadOnly() && poison != 0 {
		setSomePoisonDataCall4EEA90(obj, int32(poison))
	}
	if !cf.ReadOnly() {
		return nil
	}
	if field361High != 0 && noxflags.HasGame(noxflags.GameHost) {
		obj.HealthData.Cur = 0
		obj.HealthData.Field2 = 0
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
	if shop.SellMultiplier, err = monsterRWF32(cf, shop.SellMultiplier); err != nil {
		return fmt.Errorf("shop sell multiplier: %w", err)
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
	if !cf.ReadOnly() {
		if typ := srv.Types.ByInd(int(item.TypeInd)); typ != nil {
			typeName = typ.ID()
		}
	}
	typeName, err = monsterRWString8(cf, typeName)
	if err != nil {
		return err
	}
	if cf.ReadOnly() {
		typ := srv.Types.ByID(typeName)
		if typ == nil {
			return fmt.Errorf("unknown object type %q", typeName)
		}
		item.TypeInd = uint32(typ.Ind())
	}
	if version >= 47 {
		paramName := ""
		if !cf.ReadOnly() && item.Param != 0 {
			paramName = spell.ID(item.Param).String()
		}
		paramName, err = monsterRWString8(cf, paramName)
		if err != nil {
			return err
		}
		if cf.ReadOnly() {
			id, e := monsterParseSpellID528DB0(paramName, true)
			if e != nil {
				return fmt.Errorf("shop spell: %w", e)
			}
			item.Param = uint32(id)
		}
	}
	for i := range item.ModifierSlots {
		if !cf.ReadOnly() && item.ModifierSlots[i] != 0 {
			return fmt.Errorf("native-width modifier slot %d is not restored", i)
		}
		if _, err = monsterRWString8(cf, ""); err != nil {
			return err
		}
		if cf.ReadOnly() {
			item.ModifierSlots[i] = 0
		}
	}
	return nil
}

func monsterXferVoice528DB0(cf *cryptfile.CryptFile, ud *server.MonsterUpdateData) error {
	name := ""
	if !cf.ReadOnly() {
		name = monsterSoundSetName528DB0(ud.SoundSet122)
	}
	name, err := monsterRWString8(cf, name)
	if err != nil {
		return err
	}
	if cf.ReadOnly() {
		ud.SoundSet122 = getDefaultSoundSet(name)
	}
	return nil
}

func monsterSoundSetName528DB0(data unsafe.Pointer) string {
	if data == nil {
		return ""
	}
	for name, rec := range soundSetByName {
		if rec != nil && rec.data == data {
			return name
		}
	}
	return ""
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

func monsterOnSpawnNative529BC0(obj *server.Object) {
	if obj == nil || obj.HealthData == nil {
		return
	}
	ud := obj.UpdateDataMonster()
	def := Nox_xxx_monsterDefByTT_517560(int(obj.TypeInd))
	if def == nil {
		return
	}
	if byte(ud.Field361>>8) == 0 {
		health := uint16(def.Health68)
		obj.HealthData.Cur = health
		obj.HealthData.Field2 = health
		obj.HealthData.Max = health
	}
	if byte(ud.Field361) == 1 {
		ud.StatusFlags = def.StatusFlags92
	} else {
		const mutable = object.MonsterStatus((1<<22)-1) &^ object.MonsterStatus(0x19c40)
		ud.StatusFlags = ud.StatusFlags&^mutable | def.StatusFlags92&mutable
	}
	if ud.StatusFlags&0x20 == 0 {
		ud.StatusFlags &^= 0x1800
	}
	if byte(ud.Field335) == 1 {
		ud.RetreatLevel = def.RetreatRatio80
	}
	if byte(ud.Field337) == 1 {
		ud.ResumeLevel = def.ResumeRatio84
	}
	if byte(ud.Field509) == 1 {
		obj.Server().MonsterAutoSpells54C0C0(obj)
	}
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

func monsterResolveXferRefs528DB0(obj *server.Object) {
	if obj == nil || obj.UpdateData == nil {
		return
	}
	monsterXferPending528DB0.Lock()
	refs := monsterXferPending528DB0.m[obj]
	delete(monsterXferPending528DB0.m, obj)
	monsterXferPending528DB0.Unlock()
	if refs == nil {
		return
	}
	applyMonsterXferRefs528DB0(obj, refs)
}

func monsterObjectByNetCode528DB0(srv *server.Server, code uint32) *server.Object {
	if srv == nil || srv.ObjectByNetCode == nil || code == 0 {
		return nil
	}
	return srv.ObjectByNetCode(int(code))
}

func monsterActionByName528DB0(name string, limit int, fallback ai.ActionType) ai.ActionType {
	for i := 0; i < limit; i++ {
		action := ai.ActionType(i)
		if action.String() == name {
			return action
		}
	}
	return fallback
}

func monsterShiftFrame528DB0(value uint32, delta int32) uint32 {
	shifted := int32(value + uint32(delta))
	if shifted < 1 {
		return 1
	}
	return uint32(shifted)
}

func monsterRWBytes528DB0(cf *cryptfile.CryptFile, p []byte) error {
	if len(p) == 0 {
		return nil
	}
	if cf.ReadOnly() {
		_, err := io.ReadFull(cf, p)
		return err
	}
	n, err := cf.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	return err
}

func monsterRWU8(cf *cryptfile.CryptFile, value byte) (byte, error) {
	buf := [1]byte{value}
	err := monsterRWBytes528DB0(cf, buf[:])
	return buf[0], err
}

func monsterRWU16(cf *cryptfile.CryptFile, value uint16) (uint16, error) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], value)
	err := monsterRWBytes528DB0(cf, buf[:])
	return binary.LittleEndian.Uint16(buf[:]), err
}

func monsterRWU32(cf *cryptfile.CryptFile, value uint32) (uint32, error) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	err := monsterRWBytes528DB0(cf, buf[:])
	return binary.LittleEndian.Uint32(buf[:]), err
}

func monsterRWI32(cf *cryptfile.CryptFile, value int32) (int32, error) {
	v, err := monsterRWU32(cf, uint32(value))
	return int32(v), err
}

func monsterRWF32(cf *cryptfile.CryptFile, value float32) (float32, error) {
	v, err := monsterRWU32(cf, math.Float32bits(value))
	return math.Float32frombits(v), err
}

func monsterRWString8(cf *cryptfile.CryptFile, value string) (string, error) {
	if len(value) > math.MaxUint8 {
		value = value[:math.MaxUint8]
	}
	sz, err := monsterRWU8(cf, byte(len(value)))
	if err != nil {
		return "", err
	}
	if cf.ReadOnly() {
		buf := make([]byte, int(sz))
		if err := monsterRWBytes528DB0(cf, buf); err != nil {
			return "", err
		}
		return string(buf), nil
	}
	if err := monsterRWBytes528DB0(cf, []byte(value)); err != nil {
		return "", err
	}
	return value, nil
}

func cStringBytes528DB0(buf []byte) string {
	for i, b := range buf {
		if b == 0 {
			return string(buf[:i])
		}
	}
	return string(buf)
}

func setCStringBytes528DB0(buf []byte, value string) {
	clear(buf)
	if len(buf) == 0 {
		return
	}
	copy(buf[:len(buf)-1], value)
}

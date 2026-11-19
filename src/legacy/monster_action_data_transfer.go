package legacy

/*
#include "GAME4_2.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/common/unit/ai"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

const monsterActionXferVersion529CE0 = 4

//export nox_xxx_XFer_ActionData_529CE0
func nox_xxx_XFer_ActionData_529CE0(objC *nox_object_t) C.int {
	obj := asObjectS(objC)
	cf := cryptfile.Global()
	if cf == nil || obj == nil || obj.UpdateData == nil {
		return 0
	}
	if cf.ReadOnly() {
		obj.TakeMonsterActionRefs()
	}
	refs := new(server.MonsterActionRefs)
	if err := monsterXferActionData529CE0(cf, obj, refs, true); err != nil {
		mapLog.Printf("nox_xxx_XFer_ActionData_529CE0: %v", err)
		return 0
	}
	if cf.ReadOnly() {
		obj.StoreMonsterActionRefs(refs)
	}
	return 1
}

func monsterRWActionWordPair529CE0(cf *cryptfile.CryptFile, left, right *uint32) error {
	var buf [8]byte
	binary.LittleEndian.PutUint32(buf[:4], *left)
	binary.LittleEndian.PutUint32(buf[4:], *right)
	if err := monsterRWBytes528DB0(cf, buf[:]); err != nil {
		return err
	}
	if cf.ReadOnly() {
		*left = binary.LittleEndian.Uint32(buf[:4])
		*right = binary.LittleEndian.Uint32(buf[4:])
	}
	return nil
}

func monsterRWActionFloatPair529CE0(cf *cryptfile.CryptFile, x, y *float32) error {
	left, right := math.Float32bits(*x), math.Float32bits(*y)
	if err := monsterRWActionWordPair529CE0(cf, &left, &right); err != nil {
		return err
	}
	if cf.ReadOnly() {
		*x, *y = math.Float32frombits(left), math.Float32frombits(right)
	}
	return nil
}

// The whole path is one item in the original CRC stream. An empty path
// still writes one zero-byte item, just like an empty action string.
func monsterRWActionPath529CE0(cf *cryptfile.CryptFile, path []types.Pointf) error {
	if len(path) == 0 && !cf.ReadOnly() {
		_, err := cf.Write(nil)
		return err
	}
	var buf [8 * len(server.MonsterUpdateData{}.Path)]byte
	data := buf[:8*len(path)]
	if !cf.ReadOnly() {
		for i, point := range path {
			binary.LittleEndian.PutUint32(data[8*i:], math.Float32bits(point.X))
			binary.LittleEndian.PutUint32(data[8*i+4:], math.Float32bits(point.Y))
		}
	}
	if err := monsterRWBytes528DB0(cf, data); err != nil {
		return err
	}
	if cf.ReadOnly() {
		for i := range path {
			path[i].X = math.Float32frombits(binary.LittleEndian.Uint32(data[8*i:]))
			path[i].Y = math.Float32frombits(binary.LittleEndian.Uint32(data[8*i+4:]))
		}
	}
	return nil
}

func monsterXferActionData529CE0(cf *cryptfile.CryptFile, obj *server.Object, refs *server.MonsterActionRefs, clampOnWrite bool) error {
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
		if enabled == 0 {
			return nil
		}
	}

	frame := obj.Server().Frame()
	storedFrame, err := monsterRWU32(cf, frame)
	if err != nil {
		return fmt.Errorf("action frame: %w", err)
	}
	delta := int32(frame - storedFrame)
	// The original C entry also clamps live deadlines after writing them.
	clampDeadlines := cf.ReadOnly() || clampOnWrite
	if _, err = monsterRWU32(cf, 0); err != nil {
		return fmt.Errorf("action scratch: %w", err)
	}
	if ud.Field2, err = monsterRWU32(cf, ud.Field2); err != nil {
		return fmt.Errorf("path count: %w", err)
	}
	if ud.Field2 > uint32(len(ud.Path)) {
		return fmt.Errorf("path count %d exceeds %d", ud.Field2, len(ud.Path))
	}
	if err := monsterRWActionPath529CE0(cf, ud.Path[:ud.Field2]); err != nil {
		return fmt.Errorf("path: %w", err)
	}
	if ud.Field67, err = monsterRWU32(cf, ud.Field67); err != nil {
		return fmt.Errorf("field67: %w", err)
	}
	if err := monsterRWActionFloatPair529CE0(cf, &ud.Field68.X, &ud.Field68.Y); err != nil {
		return fmt.Errorf("field68: %w", err)
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
	if ud.Field91, err = monsterRWU32(cf, ud.Field91); err != nil {
		return fmt.Errorf("field91: %w", err)
	}
	if err := monsterRWActionWordPair529CE0(cf, &ud.Field92, &ud.Field93); err != nil {
		return fmt.Errorf("field92 and field93: %w", err)
	}
	if ud.Direction94, err = monsterRWU32(cf, ud.Direction94); err != nil {
		return fmt.Errorf("direction94: %w", err)
	}
	if err := monsterRWActionFloatPair529CE0(cf, &ud.Pos95.X, &ud.Pos95.Y); err != nil {
		return fmt.Errorf("pos95: %w", err)
	}
	soundName := ""
	if !cf.ReadOnly() {
		soundName = sound.ID(ud.Field97).String()
	}
	soundName, err = monsterRWString8(cf, soundName)
	if err != nil {
		return fmt.Errorf("sound: %w", err)
	}
	if cf.ReadOnly() {
		ud.Field97 = uint32(sound.ByName(soundName))
	}
	if err := monsterRWActionFloatPair529CE0(cf, &ud.Field99X, &ud.Field99Y); err != nil {
		return fmt.Errorf("field99: %w", err)
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
	if version < 3 {
		if _, err := monsterRWU32(cf, 0); err != nil {
			return fmt.Errorf("version %d compatibility word: %w", version, err)
		}
	}
	if ud.Field124, err = monsterRWU32(cf, ud.Field124); err != nil {
		return fmt.Errorf("field124: %w", err)
	}
	if err := monsterRWActionWordPair529CE0(cf, &ud.Field125, &ud.Field126); err != nil {
		return fmt.Errorf("field125 and field126: %w", err)
	}
	if cf.ReadOnly() {
		ud.Field124 = monsterShiftFrame528DB0(ud.Field124, delta)
	}
	if _, err = monsterRWU32(cf, ud.Field74); err != nil {
		return fmt.Errorf("waypoint loop scratch: %w", err)
	}
	if ud.Field134, err = monsterRWU32(cf, ud.Field134); err != nil {
		return fmt.Errorf("field134: %w", err)
	}
	if ud.Field135, err = monsterRWU32(cf, ud.Field135); err != nil {
		return fmt.Errorf("field135: %w", err)
	}
	if clampDeadlines {
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
		if err := monsterXferAIStackItem52A440(cf, &ud.AIStack[i], &refs.AIObjects[i], &refs.AIWaypoints[i], delta); err != nil {
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
			refs.Seen[i] = id
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
		refs.Current = current
		ud.CurrentEnemy = nil
	}
	if ud.Field301, err = monsterRWU32(cf, ud.Field301); err != nil {
		return fmt.Errorf("field301: %w", err)
	}
	if clampDeadlines {
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
	return monsterXferActionDataV4Tail529CE0(cf, obj, refs, delta, clampDeadlines)
}

func monsterXferAIStackItem52A440(cf *cryptfile.CryptFile, item *server.AIStackItem, objectIDs *[2]int32, waypointIDs *[2]uint32, delta int32) error {
	actionName := ai.ActionType(item.Action).String()
	// Slot 39 is unnamed in the original wire table. Keep String's
	// diagnostic spelling separate from the serialized action name.
	if name, ok := monsterActionNameNative534650(item.Action); ok {
		actionName = name
	}
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
		if actionName == "" {
			item.Action = 39
		}
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
			// A PE32 vector is one eight-byte I/O/CRC item, not two
			// scalar items or two native-width uintptrs.
			var buf [8]byte
			binary.LittleEndian.PutUint32(buf[:4], uint32(item.Args[slot]))
			binary.LittleEndian.PutUint32(buf[4:], uint32(item.Args[slot+1]))
			var n int
			if cf.ReadOnly() {
				n, err = io.ReadFull(cf, buf[:])
				if n >= 4 {
					item.Args[slot] = uintptr(binary.LittleEndian.Uint32(buf[:4]))
				}
				if n == 4 && err == io.ErrUnexpectedEOF {
					// Preserve the native reader's complete-word commit
					// and second-word EOF error, without admitting CRC
					// for an incomplete eight-byte item.
					err = io.EOF
				}
			} else {
				n, err = cf.Write(buf[:])
				if err == nil && n != len(buf) {
					err = io.ErrShortWrite
				}
			}
			if err != nil {
				return fmt.Errorf("argument %d.%d: %w", i, min(n/4, 1), err)
			}
			item.Args[slot] = uintptr(binary.LittleEndian.Uint32(buf[:4]))
			item.Args[slot+1] = uintptr(binary.LittleEndian.Uint32(buf[4:]))
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
				ind = legacyActionWaypointIndex(item.Args[slot])
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

func monsterXferActionDataV4Tail529CE0(cf *cryptfile.CryptFile, obj *server.Object, refs *server.MonsterActionRefs, delta int32, clampDeadlines bool) error {
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
		refs.Field98 = field98ID
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
		if clampDeadlines && it.frame {
			*it.p = monsterShiftFrame528DB0(*it.p, delta)
		}
	}
	if ud.Stamina, err = monsterRWU8(cf, ud.Stamina); err != nil {
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
		refs.Field300 = field300ID
		ud.Field300 = 0
	}
	if ud.Field302, err = monsterRWU32(cf, ud.Field302); err != nil {
		return fmt.Errorf("field302: %w", err)
	}
	if ud.Field303, err = monsterRWU32(cf, ud.Field303); err != nil {
		return fmt.Errorf("field303: %w", err)
	}
	if clampDeadlines {
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
		refs.Preferred = preferred
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
			refs.Field535[i] = id
			ud.Field535[i] = 0
		}
	}
	return nil
}

func monsterObjectByNetCode528DB0(srv *server.Server, code uint32) *server.Object {
	if code == 0 {
		return nil
	}
	return srv.ObjectFromNetCode4ECCB0(code)
}

func monsterShiftFrame528DB0(value uint32, delta int32) uint32 {
	return uint32(max(shiftedMonsterActionFrame(value, delta), 1))
}

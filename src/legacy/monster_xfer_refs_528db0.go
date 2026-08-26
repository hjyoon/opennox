package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// monsterXferRefs528DB0 keeps PE32 script/waypoint identifiers out of native
// pointer fields between the object pass and the map's deferred-reference
// pass. Storing those identifiers as pointers is invalid on every 64-bit
// target and was the source of the original truncation crash.
type monsterXferRefs528DB0 struct {
	seen      [16]int32
	current   int32
	preferred int32
	field98   int32
	field300  int32
	field535  [8]int32
	aiObjects [24][2]int32
	aiWPs     [24][2]uint32
}

func applyMonsterXferRefs528DB0(obj *server.Object, refs *monsterXferRefs528DB0) {
	if obj == nil || obj.UpdateData == nil || refs == nil {
		return
	}
	ud := obj.UpdateDataMonster()
	srv := obj.Server()
	for i := 0; i < int(ud.Field282_1); i++ {
		ud.SeenEnemies[i] = srv.ObjectByScriptID4ECF10(refs.seen[i])
		if ud.SeenEnemies[i] == nil {
			ud.Field282_1 = 0
		}
	}
	ud.CurrentEnemy = srv.ObjectByScriptID4ECF10(refs.current)
	if int32(obj.ObjFlags) >= 0 && ud.AIStackInd >= 0 {
		for i := 0; i <= int(ud.AIStackInd); i++ {
			item := &ud.AIStack[i]
			action := int(item.Action)
			if action < 0 || action >= 72 {
				continue
			}
			argCount := int(memmap.Uint8(0x587000, uintptr(255604+16*action)))
			if argCount > 2 {
				argCount = 2
			}
			for j := 0; j < argCount; j++ {
				slot := 2 * j
				switch memmap.Uint32(0x587000, uintptr(255608+4*(j+4*action))) {
				case 1:
					item.Args[slot] = uintptr(unsafe.Pointer(srv.ObjectByScriptID4ECF10(refs.aiObjects[i][j])))
				case 2:
					item.Args[slot] = uintptr(unsafe.Pointer(srv.WPs.ByInd(int(refs.aiWPs[i][j]))))
				}
			}
		}
	}
	if target := srv.ObjectByScriptID4ECF10(refs.field98); target != nil {
		ud.Field98 = target.NetCode
	} else {
		ud.Field98 = 0
	}
	if target := srv.ObjectByScriptID4ECF10(refs.field300); target != nil {
		ud.Field300 = target.NetCode
	} else {
		ud.Field300 = 0
	}
	ud.PreferredEnemy = srv.ObjectByScriptID4ECF10(refs.preferred)
	for i := 0; i < int(ud.Field543_0); i++ {
		if target := srv.ObjectByScriptID4ECF10(refs.field535[i]); target != nil {
			ud.Field535[i] = target.NetCode
		} else {
			ud.Field543_0 = 0
		}
	}
}

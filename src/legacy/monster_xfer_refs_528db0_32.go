//go:build 386 || arm

package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// resolveLegacyMonsterRefs528DB0 converts the temporary PE32 slots used by
// legacy 32-bit loaders that did not provide typed action references.
func resolveLegacyMonsterRefs528DB0(obj *server.Object) {
	ud := obj.UpdateDataMonster()
	refs := new(server.MonsterActionRefs)
	for i := 0; i < int(ud.Field282_1) && i < len(refs.Seen); i++ {
		refs.Seen[i] = int32(uintptr(unsafe.Pointer(ud.SeenEnemies[i])))
	}
	refs.Current = int32(uintptr(unsafe.Pointer(ud.CurrentEnemy)))
	refs.Field98 = int32(ud.Field98)
	refs.Field300 = int32(ud.Field300)
	refs.Preferred = int32(uintptr(unsafe.Pointer(ud.PreferredEnemy)))
	for i := 0; i < int(ud.Field543_0) && i < len(refs.Field535); i++ {
		refs.Field535[i] = int32(ud.Field535[i])
	}
	if int32(obj.ObjFlags) >= 0 && ud.AIStackInd >= 0 {
		for i := 0; i <= int(ud.AIStackInd) && i < len(refs.AIObjects); i++ {
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
					refs.AIObjects[i][j] = int32(item.Args[slot])
				case 2:
					refs.AIWaypoints[i][j] = uint32(item.Args[slot])
				}
			}
		}
	}
	applyMonsterXferRefs528DB0(obj, refs)
}

//go:build 386 || arm

package legacy

import (
	"unsafe"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

// monsterResolveXferRefs528DB0 converts the temporary PE32 pointer slots used
// by the original 32-bit transfer routine into fixed-width IDs before sharing
// the typed resolver with native-width targets.
func monsterResolveXferRefs528DB0(obj *server.Object) {
	if obj == nil || obj.UpdateData == nil {
		return
	}
	ud := obj.UpdateDataMonster()
	refs := new(monsterXferRefs528DB0)
	for i := 0; i < int(ud.Field282_1) && i < len(refs.seen); i++ {
		refs.seen[i] = int32(uintptr(unsafe.Pointer(ud.SeenEnemies[i])))
	}
	refs.current = int32(uintptr(unsafe.Pointer(ud.CurrentEnemy)))
	refs.field98 = int32(ud.Field98)
	refs.field300 = int32(ud.Field300)
	refs.preferred = int32(uintptr(unsafe.Pointer(ud.PreferredEnemy)))
	for i := 0; i < int(ud.Field543_0) && i < len(refs.field535); i++ {
		refs.field535[i] = int32(ud.Field535[i])
	}
	if int32(obj.ObjFlags) >= 0 && ud.AIStackInd >= 0 {
		for i := 0; i <= int(ud.AIStackInd) && i < len(refs.aiObjects); i++ {
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
					refs.aiObjects[i][j] = int32(item.Args[slot])
				case 2:
					refs.aiWPs[i][j] = uint32(item.Args[slot])
				}
			}
		}
	}
	applyMonsterXferRefs528DB0(obj, refs)
}

package legacy

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/opennox/v1/server"
)

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

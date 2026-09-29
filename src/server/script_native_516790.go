package server

import "github.com/opennox/libs/object"

// ScriptUnknownB8516790 reports bit 8 of the object's subclass dword.
// GAME.EXE 00516790 exposed this exact bit through NoxScript.
func ScriptUnknownB8516790(obj *Object) bool {
	return obj != nil && uint32(obj.ObjSubClass)&0x100 != 0
}

// ScriptUnknownB9516850 reports bit 7 of the object's subclass dword.
// GAME.EXE 00516850 exposed this exact bit through NoxScript.
func ScriptUnknownB9516850(obj *Object) bool {
	return obj != nil && uint32(obj.ObjSubClass)&0x80 != 0
}

var scriptHalberdTypeIDs516890 = [...]string{
	"OblivionHalberd",
	"OblivionHeart",
	"OblivionWierdling",
	"OblivionOrb",
}

const scriptOblivionWeaponMask516890 = object.WeaponStaffOblivionHalberd |
	object.WeaponStaffOblivionHeart |
	object.WeaponStaffOblivionWierdling |
	object.WeaponStaffOblivionOrb

// ScriptSetHalberdRuntime516890 supplies the engine operations used by the
// native-width port of GAME.EXE 00516890.
type ScriptSetHalberdRuntime516890 struct {
	DelayedDelete func(item *Object)
	Respawn       func(owner *Object, typeID string) *Object
	TryEquip      func(owner, item *Object)
}

// ScriptSetHalberd516890 replaces the host's current Oblivion weapon with the
// requested upgrade and preserves its equipped state. Invalid hosts and level
// values are ignored instead of indexing beyond the original four-entry PE32
// table.
func ScriptSetHalberd516890(host *Object, upgrade int, runtime ScriptSetHalberdRuntime516890) {
	if host == nil || upgrade < 0 || upgrade >= len(scriptHalberdTypeIDs516890) {
		return
	}
	equipped := false
	for item := host.FirstItem(); item != nil; item = item.NextItem() {
		if !item.Class().Has(object.ClassWeapon) ||
			!item.WeaponClass().Has(scriptOblivionWeaponMask516890) {
			continue
		}
		equipped = item.Flags().Has(object.FlagEquipped)
		runtime.DelayedDelete(item)
		break
	}
	item := runtime.Respawn(host, scriptHalberdTypeIDs516890[upgrade])
	if equipped {
		runtime.TryEquip(host, item)
	}
}

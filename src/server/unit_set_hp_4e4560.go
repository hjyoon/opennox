package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

type UnitSetHPRuntime4E4560 struct {
	CheckHeld   func(uint32, *Object)
	SetPlayerHP func(uint32, uint16)
	ProtectHeld func(uint32, *Object) int32
}

// SetHP4E4560 updates the live health record and sync masks in original
// order. The result carries the current holder address or the signed status
// returned by the final protection call.
func (obj *Object) SetHP4E4560(amount uint16, rt UnitSetHPRuntime4E4560) uintptr {
	if obj.HealthData == nil {
		return 0
	}
	obj.NeedSync()
	if holder := obj.InvHolder; holder != nil && holder.Class().Has(object.ClassPlayer) {
		rt.CheckHeld(holder.UpdateDataPlayer().Player.Prot4632, obj)
	}

	obj.HealthData.Cur = amount
	if obj.Class().Has(object.ClassPlayer) {
		rt.SetPlayerHP(obj.UpdateDataPlayer().Player.ProtUnitHPCur, amount)
	}

	class := obj.Class()
	if class.Has(object.ClassMonster) && uint32(obj.ObjSubClass)&0x80 != 0 {
		if class.HasAny(object.ClassClientPersist | object.ClassImmobile | object.ClassPlayer) {
			for i := range obj.Field140 {
				obj.Field140[i] = obj.Field140[i]&0xfffff000 | 0x20000
			}
		} else {
			obj.Sub_4E4500(0x20000, 2, obj.HasSyncData(2))
		}
	}

	holder := obj.InvHolder
	if holder != nil && holder.Class().Has(object.ClassPlayer) {
		status := rt.ProtectHeld(holder.UpdateDataPlayer().Player.Prot4632, obj)
		return uintptr(int64(status))
	}
	return uintptr(unsafe.Pointer(holder))
}

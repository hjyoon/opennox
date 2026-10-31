package server

import "unsafe"

// SetNPCWeaponForMeleeUpdate532130 publishes the logical Field516 pointer
// at the same point as the original equip/dequip store. Its PE32 word and
// adjacent animation field keep their original layout on every target.
func (obj *Object) SetNPCWeaponForMeleeUpdate532130(data *MonsterUpdateData, weapon *Object) {
	if data == nil {
		return
	}
	if unsafe.Sizeof(uintptr(0)) == 4 {
		data.Field516 = uint32(uintptr(unsafe.Pointer(weapon)))
	} else {
		data.Field516 = 0
	}
	if obj == nil || obj.serverHandle == 0 {
		return
	}
	ext := obj.SetExt()
	if ext.npcWeaponByUpdate532130 == nil {
		ext.npcWeaponByUpdate532130 = make(map[*MonsterUpdateData]*Object)
	}
	ext.npcWeaponByUpdate532130[data] = weapon
}

// NPCWeaponForMeleeUpdate532130 reads the caller's cached record. A nil map
// entry is an authoritative cleared pointer, distinct from an older inventory
// load for which no native association has been published yet.
func (obj *Object) NPCWeaponForMeleeUpdate532130(data *MonsterUpdateData) *Object {
	if data == nil {
		return nil
	}
	// Unregistered Go objects can carry equipped inventory before an original
	// pointer has been published. Preserve that compatibility path on 32-bit
	// too; registered owners and every nonzero PE32 reference use the word.
	if unsafe.Sizeof(uintptr(0)) == 4 &&
		(data.Field516 != 0 || obj == nil || obj.serverHandle != 0) {
		word := data.Field516
		return *(**Object)(unsafe.Pointer(&word))
	}
	if obj == nil || obj.serverHandle == 0 {
		return obj.NPCEquippedWeapon538960()
	}
	ext := obj.GetExt()
	if weapon, known := ext.npcWeaponByUpdate532130[data]; known {
		return weapon
	}
	if obj.UpdateData != unsafe.Pointer(data) {
		return nil
	}
	// Older native NPC loads reconstruct this reference from inventory. Seed
	// the entry-time association before callbacks can replace the live record.
	weapon := obj.NPCEquippedWeapon538960()
	if ext.npcWeaponByUpdate532130 == nil {
		ext.npcWeaponByUpdate532130 = make(map[*MonsterUpdateData]*Object)
	}
	ext.npcWeaponByUpdate532130[data] = weapon
	return weapon
}

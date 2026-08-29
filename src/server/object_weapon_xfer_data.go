package server

import "unsafe"

// WeaponArmorUpdateData is the exact fixed-width eight-byte record registered
// for WeaponArmorUpdate and DamageRoundoffUpdate. WeaponXfer and ArmorXfer
// transfer Field4 in their newest wire versions.
type WeaponArmorUpdateData struct {
	Field0 uint32
	Field4 uint32
}

var (
	_ = [1]struct{}{}[8-unsafe.Sizeof(WeaponArmorUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(WeaponArmorUpdateData{}.Field0)]
	_ = [1]struct{}{}[4-unsafe.Offsetof(WeaponArmorUpdateData{}.Field4)]
)

func (obj *Object) UpdateDataWeaponArmor() *WeaponArmorUpdateData {
	// Preserve the raw entry pointer and the original nil-fault boundary.
	return (*WeaponArmorUpdateData)(obj.UpdateData)
}

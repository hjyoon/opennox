package server

import "github.com/opennox/libs/object"

// PlayerDamageElectricArmorScale4E2220 preserves the original material-byte
// admission and binary32 spill after each addition. Material at PE32 offset
// 24 is independent of the armor subclass at offset 12.
func PlayerDamageElectricArmorScale4E2220(target *Object, specialCoefficient float32, definition func(int) *Modifier) float32 {
	scale := float32(1)
	for item := target.InvFirstItem; item != nil; item = item.InvNextItem {
		if !item.Class().Has(object.ClassArmor) || uint8(item.Material)&0x10 == 0 ||
			!item.Flags().Has(object.FlagEquipped) || ItemHasMaterial7Modifier4133D0(item) {
			continue
		}
		coefficient := float32(0)
		if uint32(item.ObjSubClass)&0x2000000 != 0 {
			coefficient = specialCoefficient
		} else if def := definition(int(item.TypeInd)); def != nil {
			coefficient = def.DamageCoeffOrArmor64
		}
		scale = experienceSpill32TowardZero(experienceAdd53TowardZero(float64(coefficient), float64(scale)))
	}
	return scale
}

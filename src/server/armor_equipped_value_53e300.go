package server

import "github.com/opennox/libs/object"

// EquippedArmorValue53E300 accumulates each item at the original startup
// x87 precision and spills its subtotal toward zero before the next item.
func EquippedArmorValue53E300(owner *Object, defense func(*Object) float64, cap float64) float32 {
	var value float32
	for item := owner.InvFirstItem; item != nil; item = item.InvNextItem {
		if item.ObjClass.Has(object.ClassArmor) && item.Flags().Has(object.FlagEquipped) {
			sum := objectForceAdd53_52DF80(float64(value), defense(item))
			value = experienceSpill32TowardZero(sum)
		}
	}
	if float64(value) > cap {
		return 1
	}
	return value
}

package server

// EffectContinualReplenishment4E02C0 is the equipped item's update callback
// from GAME.EXE 004E02C0. Object and modifier addresses remain native-width;
// only the fixed-width WandUseData charge byte is written. In particular the
// original wraps 255 to zero before clamping, and does not update Progress.
func EffectContinualReplenishment4E02C0(
	effect *ModifierEff, item *Object,
	frame func() uint32, report func(uint8, *Object, uint8, uint8),
) {
	if item == nil {
		return
	}
	now := frame()
	period := uint32(effect.Update100.Val)
	if now%period != 0 {
		return
	}
	class := uint32(item.ObjClass)
	// Load the pointer before the class gate, without validating or reading
	// its payload for non-wands, as in the original instructions.
	data := (*WandUseData)(item.UseData.Ptr)
	if class&0x1000 == 0 {
		return
	}
	old, maximum := data.Charge, data.MaxCharge
	data.Charge++
	if data.Charge > maximum {
		data.Charge = maximum
	}
	charge := data.Charge
	if old == charge {
		return
	}
	holder := item.InvHolder
	if holder != nil && uint32(holder.ObjClass)&4 != 0 {
		index := holder.UpdateDataPlayer().Player.PlayerInd
		report(index, item, charge, maximum)
	}
}

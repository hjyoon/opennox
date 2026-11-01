package timer

// ScalePercentageNative486640 preserves the complete raw 00486640 entry.
// The original multiplies in a wrapping 32-bit register before signed
// division by100. A nil group faults before arithmetic.
func ScalePercentageNative486640(group *TimerGroup, rate int32) int32 {
	percent := group.Timers[1].Current >> 16
	product := int32(uint32(rate) * percent)
	return product / 100
}

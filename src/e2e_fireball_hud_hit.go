package opennox

func e2eFireballHUDHit(meter e2eHUDMeterState, ready bool, before uint16, damage int32, originalMax uint16) bool {
	return ready && damage > 0 && damage < int32(before) &&
		meter.Current == uint32(before)-uint32(damage) && meter.Maximum == uint32(originalMax)
}

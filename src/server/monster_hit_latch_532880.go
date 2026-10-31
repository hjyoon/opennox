package server

// MonsterHitLatch532880 records the first successful hit's live frame and
// returns the original cached update pointer. The original entry neither
// checks class nor reads the clock when the timestamp is already nonzero.
func MonsterHitLatch532880(unit *Object, frame func() uint32) *MonsterUpdateData {
	update := (*MonsterUpdateData)(unit.UpdateData)
	if update.Field130 == 0 {
		update.Field130 = frame()
	}
	return update
}

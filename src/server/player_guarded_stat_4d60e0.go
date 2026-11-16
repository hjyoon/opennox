package server

// RecordGuardedStat4D60E0 applies GAME.EXE's Player stores after the
// caller has checked the Object destroyed flag. Other gate values keep
// both original counters unchanged.
func (p *Player) RecordGuardedStat4D60E0() {
	if p.Field4792 == 1 {
		p.field4652++
		p.field4692 |= 1
	}
}

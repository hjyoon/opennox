package server

// monsterCanResumeAttackData545520 retains the original FILD/FIDIV at
// gameplay PC53/ToZero. Health is unsigned 16-bit; maximum zero supplies
// one without reading Cur. FCOMP's C0-only branch rejects unordered.
func monsterCanResumeAttackData545520(health *HealthData, update *MonsterUpdateData) bool {
	portion := float64(1)
	if maximum := health.Max; maximum != 0 {
		portion = monsterMoveForceDivChop53_50D581(float64(health.Cur), float64(maximum))
	}
	return portion >= float64(update.ResumeLevel)
}

// MonsterCanResumeAttack545520 is the raw public entry. Load both record
// pointers before health access, without adding a class or null gate.
func (unit *Object) MonsterCanResumeAttack545520() bool {
	return monsterCanResumeAttackData545520(unit.HealthData, (*MonsterUpdateData)(unit.UpdateData))
}

func monsterCanCast2Data5455B0(unit *Object, update *MonsterUpdateData) bool {
	return uint8(update.StatusFlags)&0x20 != 0 && unit.UnitBuffTest4FF350(29) != 0
}

// MonsterCanCast2_5455B0 retains the original low-byte flag gate before
// calling the shared native buff getter. Health is never accessed here.
func (unit *Object) MonsterCanCast2_5455B0() bool {
	return monsterCanCast2Data5455B0(unit, (*MonsterUpdateData)(unit.UpdateData))
}

// MonsterRetreatActive545580 retains the cached unit and original child
// short circuit: below the resume threshold, CanCast2 is not evaluated.
func (unit *Object) MonsterRetreatActive545580() bool {
	return !unit.MonsterCanResumeAttack545520() || unit.MonsterCanCast2_5455B0()
}

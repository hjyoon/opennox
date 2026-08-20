package server

const (
	unitGiveXPScaleBits4EF270 = uint32(0x3c23d70a)
	unitGiveXPOneBits4EF270   = uint32(0x3f800000)
	unitGiveXPZeroBits4EF270  = uint32(0x00000000)
)

type unitGiveXPHooks4EF270[O, U, P any] struct {
	loadUnitArg         func() O
	loadExperience      func(O) float32
	loadTargetArg       func() float32
	loadUpdateData      func(O) U
	loadScale           func() float32
	loadOne             func() float32
	loadZero            func() float32
	storeExperience     func(O, float32)
	loadPlayer          func(U) P
	loadExperienceToken func(P) uint32
	protectExperience   func(uint32, float32)
	reportExperience    func(O)
	syncLevel           func(O)
}

// GAME.EXE runs this routine under x87 53-bit precision. Keep every arithmetic
// boundary explicit so a host compiler cannot contract or reassociate it.
//
//go:noinline
func unitGiveXPSub64_4EF270(a, b float64) float64 { return a - b }

//go:noinline
func unitGiveXPMul64_4EF270(a, b float64) float64 { return a * b }

//go:noinline
func unitGiveXPAdd64_4EF270(a, b float64) float64 { return a + b }

//go:noinline
func unitGiveXPSpill32_4EF270(value float64) float32 { return float32(value) }

// unitGiveXP4EF270 preserves GAME.EXE 004EF270. The initial comparison reads
// Experience before the target and returns exact positive zero for every
// ordered Experience >= target result. Unordered comparisons take the award
// path. That path reloads target and Experience, caches UpdateData, computes
// (target-Experience)*0.01f+1.0f at binary64/x87-53 boundaries, and spills a
// binary32 award without popping the retained value. The retained value is
// added to a third live Experience read and rounded only at the field store.
// Player and its protection token are reached through the cached UpdateData
// after that store. Protection, network report, and level synchronization run
// in order, and the saved binary32 award is returned after all callbacks.
func unitGiveXP4EF270[O, U, P any](hooks unitGiveXPHooks4EF270[O, U, P]) float64 {
	unit := hooks.loadUnitArg()
	experience := hooks.loadExperience(unit)
	target := hooks.loadTargetArg()
	if experience >= target {
		return float64(hooks.loadZero())
	}

	target = hooks.loadTargetArg()
	experience = hooks.loadExperience(unit)
	difference := unitGiveXPSub64_4EF270(float64(target), float64(experience))
	update := hooks.loadUpdateData(unit)
	scaled := unitGiveXPMul64_4EF270(difference, float64(hooks.loadScale()))
	adjusted := unitGiveXPAdd64_4EF270(scaled, float64(hooks.loadOne()))
	award := unitGiveXPSpill32_4EF270(adjusted)
	liveExperience := hooks.loadExperience(unit)
	newExperience := unitGiveXPAdd64_4EF270(adjusted, float64(liveExperience))
	hooks.storeExperience(unit, unitGiveXPSpill32_4EF270(newExperience))
	player := hooks.loadPlayer(update)
	token := hooks.loadExperienceToken(player)
	hooks.protectExperience(token, award)
	hooks.reportExperience(unit)
	hooks.syncLevel(unit)
	return float64(award)
}

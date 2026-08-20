package server

import "math"

const (
	directExperienceMessageKey4EF3A0  = "health.c:gainpoints"
	directExperienceMessagePath4EF3A0 = `C:\NoxPost\src\Server\GameMech\explevel.c`
	directExperienceMessageLine4EF3A0 = 381
)

type directExperienceGrantHooks4EF3A0[O, U, P, M any] struct {
	loadAwardArg        func() float32
	loadUnitArg         func() O
	loadExperience      func(O) float32
	loadUpdateData      func(O) U
	storeExperience     func(O, float32)
	loadPlayer          func(U) P
	loadExperienceToken func(P) uint32
	protectExperience   func(uint32, float32)
	reportExperience    func(O)
	loadString          func(string, string, int) M
	sendLineMessage     func(O, M, uint32)
	syncLevel           func(O)
}

// GAME.EXE performs the addition in x87 extended precision and rounds only at
// the binary32 Experience store. Both inputs originate as binary32 values, so
// binary64 preserves the required intermediate before the explicit spill.
//
//go:noinline
func directExperienceAdd64_4EF3A0(a, b float64) float64 { return a + b }

//go:noinline
func directExperienceSpill32_4EF3A0(value float64) float32 { return float32(value) }

// directExperienceTruncLow4EF3A0 models GAME.EXE 00566DCC: truncate the x87
// value toward zero into a signed qword and retain EAX, its low 32 bits. An
// invalid qword conversion produces 0x8000000000000000 and therefore zero in
// EAX.
func directExperienceTruncLow4EF3A0(value float64) uint32 {
	if math.IsNaN(value) || value >= 0x1p63 || value < -0x1p63 {
		return 0
	}
	return uint32(int64(math.Trunc(value)))
}

// directExperienceGrant4EF3A0 preserves GAME.EXE 004EF3A0. The original
// binary32 award is read before Unit, then read again as raw protection data
// before live Experience. After the x87 addition, UpdateData is cached before
// the rounded Experience store; Player and its token are reached through that
// cached UpdateData only after the store. Protection and experience reporting
// run in order.
//
// The award argument is reloaded after reporting, converted through the x87
// signed-qword helper, and passed as an unsigned low dword to the unconditional
// gainpoints lookup and line message. Level synchronization is always last.
// There are no nil, class, sign, finite-value, or zero guards in this routine.
func directExperienceGrant4EF3A0[O, U, P, M any](
	hooks directExperienceGrantHooks4EF3A0[O, U, P, M],
) {
	awardForSum := hooks.loadAwardArg()
	unit := hooks.loadUnitArg()
	awardForProtection := hooks.loadAwardArg()
	experience := hooks.loadExperience(unit)
	newExperience := directExperienceAdd64_4EF3A0(
		float64(awardForSum),
		float64(experience),
	)
	update := hooks.loadUpdateData(unit)
	hooks.storeExperience(unit, directExperienceSpill32_4EF3A0(newExperience))
	player := hooks.loadPlayer(update)
	token := hooks.loadExperienceToken(player)
	hooks.protectExperience(token, awardForProtection)
	hooks.reportExperience(unit)

	awardForMessage := hooks.loadAwardArg()
	points := directExperienceTruncLow4EF3A0(float64(awardForMessage))
	message := hooks.loadString(
		directExperienceMessageKey4EF3A0,
		directExperienceMessagePath4EF3A0,
		directExperienceMessageLine4EF3A0,
	)
	hooks.sendLineMessage(unit, message, points)
	hooks.syncLevel(unit)
}

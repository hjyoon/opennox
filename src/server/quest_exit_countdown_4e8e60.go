package server

import "math"

const (
	questExitTimerBalance4E8E60 = "QuestExitTimerStart"
	questExitCountdownID4E8E60  = "objcoll.c:ExitCountdown"
	questExitRecipient4E8E60    = int32(255)
)

type questExitCountdownHooks4E8E60[O comparable, U, P any] struct {
	balanceFloat         func(string) float64
	floatToInt           func(float32) int32
	timerActive          func() int32
	timerRemainingMillis func() int32
	firstUnit            func() O
	nextUnit             func(O) O
	loadUpdateData       func(O) U
	loadPlayer           func(U) P
	loadQuestState       func(P) uint32
	loadQuestExit        func(U) O
	stopTimer            func(int32) int32
	countdownStarted     func() int32
	startCountdown       func(int32, string)
	sendGauntlet         func(int32) int32
}

// questExitCountdown4E8E60 preserves GAME.EXE 004E8E60. It derives the
// countdown from the exact-one Quest players and the subset standing in a
// Quest exit. The balance value is first spilled to binary32, an active timer
// replaces it with signed milliseconds/1000, and the proportional result is
// rounded with the original x87 FISTP semantics.
func questExitCountdown4E8E60[O comparable, U, P any](
	hooks questExitCountdownHooks4E8E60[O, U, P],
) int32 {
	seconds := hooks.floatToInt(float32(hooks.balanceFloat(questExitTimerBalance4E8E60)))
	if hooks.timerActive() != 0 {
		seconds = hooks.timerRemainingMillis() / 1000
	}

	var zero O
	unit := hooks.firstUnit()
	if unit == zero {
		return hooks.stopTimer(0)
	}

	var total int32
	var ready int32
	for unit != zero {
		update := hooks.loadUpdateData(unit)
		player := hooks.loadPlayer(update)
		if hooks.loadQuestState(player) == 1 {
			total++
			if hooks.loadQuestExit(update) != zero {
				ready++
			}
		}
		unit = hooks.nextUnit(unit)
	}
	if total == 0 {
		return hooks.stopTimer(0)
	}

	ratio := questExitDiv64_4E8E60(float64(ready), float64(total))
	portion := questExitMul64_4E8E60(ratio, float64(seconds))
	scaled := hooks.floatToInt(float32(portion))
	candidate := seconds - scaled
	if candidate < seconds {
		seconds = candidate
	} else {
		result := hooks.countdownStarted()
		if result != 0 {
			return result
		}
	}

	hooks.startCountdown(seconds, questExitCountdownID4E8E60)
	return hooks.sendGauntlet(questExitRecipient4E8E60)
}

//go:noinline
func questExitDiv64_4E8E60(left, right float64) float64 {
	return left / right
}

//go:noinline
func questExitMul64_4E8E60(left, right float64) float64 {
	return left * right
}

// questExitRound4E8E60 models nox_float2int under the default x87
// round-to-nearest-even mode. Invalid conversions produce integer-indefinite.
func questExitRound4E8E60(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

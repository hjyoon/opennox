package server

import "math"

const (
	goldInitScaledRandomPath4F04B0 = `C:\NoxPost\src\Server\Object\init\Init.c`
	goldInitScaledRandomLine4F04B0 = int32(1017)
	goldInitBaseRandomPath4F04B0   = `C:\NoxPost\src\Server\Object\init\Init.c`
	goldInitBaseRandomLine4F04B0   = int32(1018)

	goldInitLowerScaleBits4F04B0    = uint64(0x3f847ae147ae147b)
	goldInitUpperScaleBits4F04B0    = uint64(0x3f947ae147ae147b)
	goldInitNegativeScaleBits4F04B0 = uint64(0xbf947ae147ae147b)
)

type goldInitHooks4F04B0[O, D, P comparable] struct {
	loadUnitArg    func() (O, int32)
	loadInitData   func(O) D
	loadAmount     func(D) uint32
	firstPlayer    func() P
	loadPlayerUnit func(P) O
	loadExperience func(O) float32
	nextPlayer     func(P) P
	truncQwordLow  func(float64) int32
	randomInt      func(int32, int32, string, int32) int32
	storeAmount    func(D, uint32)
}

// Keep every original x87 arithmetic boundary explicit. GAME.EXE runs this
// code with 53-bit precision control: binary32 operands are widened exactly,
// each operation is rounded as binary64, and only FSTP sites spill to
// binary32.
//
//go:noinline
func goldInitAdd64_4F04B0(left, right float64) float64 { return left + right }

//go:noinline
func goldInitDiv64_4F04B0(left, right float64) float64 { return left / right }

//go:noinline
func goldInitMul64_4F04B0(left, right float64) float64 { return left * right }

//go:noinline
func goldInitSpill32_4F04B0(value float64) float32 { return float32(value) }

func goldInitAddExperience4F04B0(sum, experience float32) float32 {
	return goldInitSpill32_4F04B0(
		goldInitAdd64_4F04B0(float64(sum), float64(experience)),
	)
}

func goldInitAverage4F04B0(sum float32, count int32) float32 {
	return goldInitSpill32_4F04B0(
		goldInitDiv64_4F04B0(float64(sum), float64(count)),
	)
}

func goldInitScale4F04B0(average float32, scaleBits uint64) float64 {
	return goldInitMul64_4F04B0(float64(average), math.Float64frombits(scaleBits))
}

func goldInitTruncQwordLow4F04B0(value float64) int32 {
	return x87TruncSignedQwordLow566DCC(value)
}

// goldInit4F04B0 preserves GAME.EXE 004F04B0's observable order. The unit
// argument and InitData pointer are cached at entry. A nonzero Amount returns
// the entry unit's low dword without touching player or RNG services.
//
// The zero path counts every Player record, including records with no unit,
// but adds Experience only for non-nil units. Every addition and the final
// average spill to binary32. The upper truncation intentionally precedes the
// lower truncation even though the RNG receives lower first. Amount arithmetic
// wraps at 32 bits, while the full second RNG result is returned. The original
// has no nil guards.
func goldInit4F04B0[O, D, P comparable](hooks goldInitHooks4F04B0[O, D, P]) int32 {
	unit, result := hooks.loadUnitArg()
	initData := hooks.loadInitData(unit)
	if hooks.loadAmount(initData) != 0 {
		return result
	}

	var sum float32
	var count int32
	var nilObject O
	var nilPlayer P
	for player := hooks.firstPlayer(); player != nilPlayer; {
		playerUnit := hooks.loadPlayerUnit(player)
		if playerUnit != nilObject {
			sum = goldInitAddExperience4F04B0(sum, hooks.loadExperience(playerUnit))
		}
		player = hooks.nextPlayer(player)
		count++
	}

	average := goldInitAverage4F04B0(sum, count)
	upper := hooks.truncQwordLow(goldInitScale4F04B0(average, goldInitUpperScaleBits4F04B0))
	lower := hooks.truncQwordLow(goldInitScale4F04B0(average, goldInitLowerScaleBits4F04B0))
	scaledRandom := hooks.randomInt(
		lower,
		upper,
		goldInitScaledRandomPath4F04B0,
		goldInitScaledRandomLine4F04B0,
	)
	negative := hooks.truncQwordLow(goldInitScale4F04B0(average, goldInitNegativeScaleBits4F04B0))
	baseRandom := hooks.randomInt(
		15,
		30,
		goldInitBaseRandomPath4F04B0,
		goldInitBaseRandomLine4F04B0,
	)
	amount := uint32(scaledRandom) - uint32(negative) + uint32(baseRandom)
	hooks.storeAmount(initData, amount)
	return baseRandom
}

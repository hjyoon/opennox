package server

import "math"

const (
	activatePoisonDestroyedFlagLow4EE7E0 = uint8(0x02)
	activatePoisonPlayerClassLow4EE7E0   = uint8(0x04)
	activatePoisonMonsterClassLow4EE7E0  = uint8(0x02)
	activatePoisonObserverFlag4EE7E0     = uint32(0x00000001)
	activatePoisonImmuneSubClass4EE7E0   = uint32(0x00000200)
	activatePoisonBlockingEnchant4EE7E0  = uint32(23)
	activatePoisonChangedSound4EE7E0     = uint32(100)
	activatePoisonPercentBits4EE7E0      = uint32(0x42c80000)
	activatePoisonRandomMinimum4EE7E0    = int32(0)
	activatePoisonRandomMaximum4EE7E0    = int32(100)
	activatePoisonRandomLine4EE7E0       = int32(361)
	activatePoisonRandomPath4EE7E0       = "C:\\NoxPost\\src\\Server\\Object\\health.c"
	activatePoisonResistMessage4EE7E0    = "Health.c:ResistPoison"
)

type activatePoisonHooks4EE7E0[O, U, P, H any] struct {
	loadUnitArg      func() O
	loadCurrent      func(O) uint8
	loadFlagsLow     func(O) uint8
	testBuff         func(O, uint32) int32
	loadClass        func(O) uint32
	loadUpdateData   func(O) U
	loadPlayer       func(U) P
	loadPlayerFlags  func(P) uint32
	loadSubClass     func(O) uint32
	poisonProtection func(O) float64
	floatToInt       func(float32) int32
	randomInt        func(int32, int32, string, int32) int32
	priorityMessage  func(O, string, uint8)
	loadIncrementArg func() int32
	loadMaximumArg   func() int32
	setPoison        func(O, int32)
	audio            func(uint32, O, int32, uint32)
	loadHealth       func(O) H
	frame            func() uint32
	storePoisonFrame func(H, uint32)
}

// activatePoison4EE7E0 preserves GAME.EXE 004EE7E0. The current poison byte is
// read before the entry nil branch, so a native nil unit faults instead of
// reaching the nominal zero return. Buff 23 rejects only the exact result one.
// ObjClass is cached across both the Player and Monster gates. The increment
// and cap arguments are deferred until resistance RNG succeeds, and the
// zero-to-positive frame stamp reloads HealthData only after set/audio effects.
func activatePoison4EE7E0[O, U, P, H comparable](
	hooks activatePoisonHooks4EE7E0[O, U, P, H],
) int32 {
	unit := hooks.loadUnitArg()
	current := hooks.loadCurrent(unit)

	var nilUnit O
	if unit == nilUnit {
		return 0
	}
	if hooks.loadFlagsLow(unit)&activatePoisonDestroyedFlagLow4EE7E0 != 0 {
		return 0
	}
	if hooks.testBuff(unit, activatePoisonBlockingEnchant4EE7E0) == 1 {
		return 0
	}

	class := hooks.loadClass(unit)
	if uint8(class)&activatePoisonPlayerClassLow4EE7E0 != 0 {
		update := hooks.loadUpdateData(unit)
		player := hooks.loadPlayer(update)
		if hooks.loadPlayerFlags(player)&activatePoisonObserverFlag4EE7E0 == activatePoisonObserverFlag4EE7E0 {
			return 0
		}
	}
	if uint8(class)&activatePoisonMonsterClassLow4EE7E0 != 0 {
		if hooks.loadSubClass(unit)&activatePoisonImmuneSubClass4EE7E0 != 0 {
			return 0
		}
	}

	protection := hooks.poisonProtection(unit)
	threshold := hooks.floatToInt(activatePoisonScaleProtection4EE7E0(protection))
	roll := hooks.randomInt(
		activatePoisonRandomMinimum4EE7E0,
		activatePoisonRandomMaximum4EE7E0,
		activatePoisonRandomPath4EE7E0,
		activatePoisonRandomLine4EE7E0,
	)
	if roll < threshold {
		hooks.priorityMessage(unit, activatePoisonResistMessage4EE7E0, 0)
		return 0
	}

	increment := hooks.loadIncrementArg()
	maximum := hooks.loadMaximumArg()
	target := activatePoisonTarget4EE7E0(current, increment, maximum)
	if target != int32(current) {
		hooks.setPoison(unit, target)
		hooks.audio(activatePoisonChangedSound4EE7E0, unit, 0, 0)
	}

	if current == 0 && target > 0 {
		health := hooks.loadHealth(unit)
		var nilHealth H
		if health != nilHealth {
			frame := hooks.frame()
			hooks.storePoisonFrame(health, frame)
		}
	}
	return 1
}

func activatePoisonTarget4EE7E0(current uint8, increment, maximum int32) int32 {
	cur := int32(current)
	sum := cur + increment
	if sum <= maximum {
		return sum
	}
	if cur <= maximum {
		return maximum
	}
	return cur
}

// activatePoisonScaleProtection4EE7E0 models FMULS by exact binary32 100.0
// under the original x87 53-bit precision control, followed by one FSTPS.
func activatePoisonScaleProtection4EE7E0(protection float64) float32 {
	scale := math.Float32frombits(activatePoisonPercentBits4EE7E0)
	return float32(protection * float64(scale))
}

// activatePoisonRound4EE7E0 models nox_float2int's default x87 FISTP mode.
// Invalid and out-of-range inputs produce the integer-indefinite 0x80000000.
func activatePoisonRound4EE7E0(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

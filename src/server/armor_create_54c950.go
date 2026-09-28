package server

const (
	armorCreateQuestFlag54C950       = uint32(0x00001000)
	armorCreateQuestMultiplier54C950 = "QuestDurabilityMultiplier"
)

type armorCreateHooks54C950[O, D, H any] struct {
	loadTypeInd    func(O) uint16
	findDefinition func(uint16) D
	loadHealth     func(O) H
	loadDurability func(D) uint16
	storeCurrent   func(H, uint16)
	storeMaximum   func(H, uint16)
	gameFlag       func(uint32) int32
	loadBalance    func(string) float32
	loadCurrent    func(H) uint16
	loadMaximum    func(H) uint16
	floatToInt     func(float32) int32
}

// armorCreate54C950 preserves GAME.EXE 0054C950. The armor definition is
// resolved before HealthData is read, and both pointers are guarded only at
// those initial gates. Every later HealthData access is a live object reload,
// matching the original PE32 instruction stream rather than caching a Go
// pointer across stores or callbacks.
func armorCreate54C950[O any, D, H comparable](obj O, hooks armorCreateHooks54C950[O, D, H]) {
	typeInd := hooks.loadTypeInd(obj)
	definition := hooks.findDefinition(typeInd)
	var nilDefinition D
	if definition == nilDefinition {
		return
	}

	health := hooks.loadHealth(obj)
	var nilHealth H
	if health == nilHealth {
		return
	}

	durability := hooks.loadDurability(definition)
	hooks.storeCurrent(health, durability)
	health = hooks.loadHealth(obj)
	durability = hooks.loadDurability(definition)
	hooks.storeMaximum(health, durability)

	if hooks.gameFlag(armorCreateQuestFlag54C950) == 0 {
		return
	}

	multiplier := hooks.loadBalance(armorCreateQuestMultiplier54C950)
	health = hooks.loadHealth(obj)
	current := hooks.loadCurrent(health)
	currentScaled := armorCreateScale54C950(current, multiplier)
	currentValue := hooks.floatToInt(currentScaled)
	health = hooks.loadHealth(obj)
	hooks.storeCurrent(health, uint16(currentValue))

	health = hooks.loadHealth(obj)
	maximum := hooks.loadMaximum(health)
	maximumScaled := armorCreateScale54C950(maximum, multiplier)
	maximumValue := hooks.floatToInt(maximumScaled)
	health = hooks.loadHealth(obj)
	hooks.storeMaximum(health, uint16(maximumValue))
}

// armorCreateScale54C950 models FILD followed by FMULS under the executable's
// 53-bit x87 precision mode and the single FSTPS spill before nox_float2int.
func armorCreateScale54C950(value uint16, multiplier float32) float32 {
	return float32(float64(value) * float64(multiplier))
}

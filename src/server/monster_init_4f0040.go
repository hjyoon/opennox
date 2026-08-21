package server

import "math"

const (
	monsterInitSkipFlags4F0040        = uint32(0x00008020)
	monsterInitNPCSubclassMask4F0040  = uint8(0x30)
	monsterInitStatusHold4F0040       = uint32(0x00000040)
	monsterInitStatusAlert4F0040      = uint32(0x00000100)
	monsterInitStatusRunning4F0040    = uint32(0x00004000)
	monsterInitStatusAlwaysRun4F0040  = uint32(0x00008000)
	monsterInitActionIdle4F0040       = uint32(0)
	monsterInitActionEscort4F0040     = uint32(3)
	monsterInitActionGuard4F0040      = uint32(4)
	monsterInitActionHunt4F0040       = uint32(5)
	monsterInitActionRoam4F0040       = uint32(10)
	monsterInitActionFight4F0040      = uint32(15)
	monsterInitActionRandomWalk4F0040 = uint32(29)
	monsterInitActionInvalid4F0040    = uint32(38)

	monsterInitAggressionBits4F0040 = uint32(0x3e23d70a)
	monsterInitSightAddBits4F0040   = uint32(0x41200000)
	monsterInitSpeedHalfBits4F0040  = uint32(0x3f000000)
	monsterInitSpeedAddBits4F0040   = uint32(0x3fd9999a)
	monsterInitSpeedMinBits4F0040   = uint32(0x3f733333)
	monsterInitSpeedMaxBits4F0040   = uint32(0x3f866666)
	monsterInitFleeRangeBits4F0040  = uint32(0x42c80000)

	monsterInitRandomSource4F0040 = `C:\NoxPost\src\Server\Object\init\Init.c`
	monsterInitRandomLine4F0040   = 0x381
)

type monsterInitHooks4F0040[O, U, H, D any, A comparable] struct {
	loadUnitArg       func() O
	loadUpdateData    func(O) U
	loadPlantTypeID   func() uint32
	loadObjectFlags   func(O) uint32
	loadTypeIndex     func(O) uint16
	isRat             func(O) bool
	isFish            func(O) bool
	isFrog            func(O) bool
	clearActionStack  func(O)
	loadMonsterDef    func(U) D
	loadMeleeRange    func(D) float32
	loadCircleRadius  func(O) float32
	loadAIAction      func(U) uint32
	storeAIAction     func(U, uint32)
	storeSightRange   func(U, float32)
	storeAggression   func(U, float32)
	loadStatus        func(U) uint32
	storeStatus       func(U, uint32)
	pushAction        func(O, uint32) A
	storeActionArg    func(A, int, uint32)
	storeActionArgLow func(A, int, uint8)
	canAttackAtWill   func(O) bool
	loadPositionXBits func(O) uint32
	loadPositionYBits func(O) uint32
	loadFrame         func() uint32
	loadDirection     func(O) int16
	storeDirection    func(U, uint32)
	storePositionX    func(U, uint32)
	storePositionY    func(U, uint32)
	loadHealth        func(O) H
	loadHealthMaximum func(H) uint16
	loadHealthCurrent func(H) uint16
	loadHealthScale   func(U) float32
	setHealth         func(O, uint16)
	storeHealthPrev   func(H, uint16)
	storeHealthGraph  func(U, int, uint16)
	loadSubclassLow   func(O) uint8
	loadSpeedField332 func(U) float32
	loadSpeedField333 func(U) uint32
	randomFloat       func(float32, float32, string, int) float64
	loadSpeedBase     func(O) float32
	storeSpeedBase    func(O, float32)
	canCast           func(O) bool
	storeFleeRange    func(U, float32)
}

// Keep GAME.EXE's x87 53-bit arithmetic boundaries explicit. All operands
// loaded by 004F0040 are binary32 or uint16, so binary64 exactly models each
// configured x87 intermediate before the original binary32 spill.
//
//go:noinline
func monsterInitAdd64_4F0040(a, b float64) float64 { return a + b }

//go:noinline
func monsterInitMul64_4F0040(a, b float64) float64 { return a * b }

//go:noinline
func monsterInitSpill32_4F0040(value float64) float32 { return float32(value) }

func monsterInitPlantSight4F0040(meleeRange, radius float32) float32 {
	value := monsterInitAdd64_4F0040(float64(meleeRange), float64(radius))
	value = monsterInitAdd64_4F0040(value, float64(math.Float32frombits(monsterInitSightAddBits4F0040)))
	return monsterInitSpill32_4F0040(value)
}

// monsterInitScaledHealth4F0040 models helper 00566DCC. The x87 product is
// truncated toward zero to a signed qword; invalid conversions yield integer
// indefinite, whose low word is zero. SetHP observes only that low word.
func monsterInitScaledHealth4F0040(maximum uint16, scale float32) uint16 {
	value := monsterInitMul64_4F0040(float64(maximum), float64(scale))
	if math.IsNaN(value) || value >= 0x1p63 || value < -0x1p63 {
		return 0
	}
	return uint16(int64(math.Trunc(value)))
}

func monsterInitNPCSpeed4F0040(field332 float32) float32 {
	value := monsterInitMul64_4F0040(
		float64(field332),
		float64(math.Float32frombits(monsterInitSpeedHalfBits4F0040)),
	)
	value = monsterInitAdd64_4F0040(
		value,
		float64(math.Float32frombits(monsterInitSpeedAddBits4F0040)),
	)
	return monsterInitSpill32_4F0040(value)
}

func monsterInitRandomSpeed4F0040(random float64, speed float32) float32 {
	return monsterInitSpill32_4F0040(
		monsterInitMul64_4F0040(random, float64(speed)),
	)
}

// monsterInit4F0040 preserves GAME.EXE 004F0040's observable load, callback,
// and store order. Unit and UpdateData are cached at entry. Type predicates,
// action services, health, speed, and capability callbacks retain their live
// loads; the 32 health-history iterations reload both HealthData and Cur.
//
// The original has no nil, class, range, or finite-value guards. The only
// returned-pointer checks are the five checks following PushAction. Numeric
// action arguments remain fixed-width dwords, while the roam flag stores only
// the low byte of argument two.
func monsterInit4F0040[O, U, H, D any, A comparable](
	h monsterInitHooks4F0040[O, U, H, D, A],
) {
	unit := h.loadUnitArg()
	update := h.loadUpdateData(unit)
	plantTypeID := h.loadPlantTypeID()
	flags := h.loadObjectFlags(unit)

	if flags&monsterInitSkipFlags4F0040 == 0 {
		typeIndex := h.loadTypeIndex(unit)
		if uint32(typeIndex) == plantTypeID {
			h.clearActionStack(unit)
			monsterDef := h.loadMonsterDef(update)
			h.storeAIAction(update, monsterInitActionGuard4F0040)
			meleeRange := h.loadMeleeRange(monsterDef)
			radius := h.loadCircleRadius(unit)
			h.storeSightRange(update, monsterInitPlantSight4F0040(meleeRange, radius))
		} else if h.isRat(unit) {
			h.clearActionStack(unit)
			h.pushAction(unit, monsterInitActionRandomWalk4F0040)
			h.storeAggression(update, math.Float32frombits(monsterInitAggressionBits4F0040))
			h.storeAIAction(update, monsterInitActionInvalid4F0040)
		} else if h.isFish(unit) {
			h.clearActionStack(unit)
			item := h.pushAction(unit, monsterInitActionRoam4F0040)
			var zero A
			if item != zero {
				h.storeActionArg(item, 0, 0)
				h.storeActionArgLow(item, 2, 0xff)
			}
			h.storeAggression(update, math.Float32frombits(monsterInitAggressionBits4F0040))
			h.storeAIAction(update, monsterInitActionInvalid4F0040)
		} else if h.isFrog(unit) {
			h.clearActionStack(unit)
			h.pushAction(unit, monsterInitActionIdle4F0040)
			status := h.loadStatus(update)
			status |= monsterInitStatusAlert4F0040
			h.storeAggression(update, math.Float32frombits(monsterInitAggressionBits4F0040))
			h.storeStatus(update, status)
			h.storeAIAction(update, monsterInitActionInvalid4F0040)
		}
	}

	action := h.loadAIAction(update)
	var zeroActionItem A
	switch action {
	case monsterInitActionFight4F0040:
		item := h.pushAction(unit, monsterInitActionFight4F0040)
		if item != zeroActionItem {
			h.storeActionArg(item, 0, h.loadPositionXBits(unit))
			h.storeActionArg(item, 1, h.loadPositionYBits(unit))
			h.storeActionArg(item, 2, h.loadFrame())
		}
	case monsterInitActionEscort4F0040:
		item := h.pushAction(unit, monsterInitActionEscort4F0040)
		if item != zeroActionItem {
			h.storeActionArg(item, 0, h.loadPositionXBits(unit))
			h.storeActionArg(item, 1, h.loadPositionYBits(unit))
			h.storeActionArg(item, 2, 0)
		}
	case monsterInitActionRoam4F0040:
		if h.canAttackAtWill(unit) {
			h.pushAction(unit, monsterInitActionHunt4F0040)
		} else {
			item := h.pushAction(unit, monsterInitActionRoam4F0040)
			if item != zeroActionItem {
				h.storeActionArg(item, 0, 0)
				h.storeActionArgLow(item, 2, uint8(h.loadSpeedField333(update)))
			}
		}
	case monsterInitActionGuard4F0040:
		item := h.pushAction(unit, monsterInitActionGuard4F0040)
		if item != zeroActionItem {
			h.storeActionArg(item, 0, h.loadPositionXBits(unit))
			h.storeActionArg(item, 1, h.loadPositionYBits(unit))
			h.storeActionArg(item, 2, uint32(int32(h.loadDirection(unit))))
		}
	case monsterInitActionInvalid4F0040:
		// No action is pushed for the original invalid sentinel.
	default:
		// Every value outside the five explicit jump-table cases reaches
		// 004F0211 and unconditionally pushes Idle.
		h.pushAction(unit, monsterInitActionIdle4F0040)
	}

	h.storeAIAction(update, monsterInitActionInvalid4F0040)
	direction := h.loadDirection(unit)
	h.storeDirection(update, uint32(int32(direction)))
	h.storePositionX(update, h.loadPositionXBits(unit))
	h.storePositionY(update, h.loadPositionYBits(unit))

	health := h.loadHealth(unit)
	maximum := h.loadHealthMaximum(health)
	current := h.loadHealthCurrent(health)
	if maximum == current {
		scale := h.loadHealthScale(update)
		h.setHealth(unit, monsterInitScaledHealth4F0040(maximum, scale))
	}
	health = h.loadHealth(unit)
	current = h.loadHealthCurrent(health)
	h.storeHealthPrev(health, current)
	for index := 0; index < 32; index++ {
		health = h.loadHealth(unit)
		current = h.loadHealthCurrent(health)
		h.storeHealthGraph(update, index, current)
	}

	if h.loadSubclassLow(unit)&monsterInitNPCSubclassMask4F0040 != 0 {
		field332 := h.loadSpeedField332(update)
		h.storeSpeedBase(unit, monsterInitNPCSpeed4F0040(field332))
	} else {
		random := h.randomFloat(
			math.Float32frombits(monsterInitSpeedMinBits4F0040),
			math.Float32frombits(monsterInitSpeedMaxBits4F0040),
			monsterInitRandomSource4F0040,
			monsterInitRandomLine4F0040,
		)
		speed := h.loadSpeedBase(unit)
		h.storeSpeedBase(unit, monsterInitRandomSpeed4F0040(random, speed))
	}

	if h.canCast(unit) {
		h.storeFleeRange(update, math.Float32frombits(monsterInitFleeRangeBits4F0040))
	}
	status := h.loadStatus(update)
	if status&monsterInitStatusHold4F0040 != 0 {
		h.storeFleeRange(update, 0)
	}
	if status&monsterInitStatusAlwaysRun4F0040 != 0 {
		status |= monsterInitStatusRunning4F0040
		h.storeStatus(update, status)
	}
}

package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// firewalkHooks52ED40 keeps object and duration identities native-width while
// preserving the original callback's frame and position state machine.
type firewalkHooks52ED40[Record, Object comparable] struct {
	loadTarget      func(Record) Object
	loadFlags       func(Object) uint32
	loadFrame60     func(Record) uint32
	loadFrame64     func(Record) uint32
	storeFrame64    func(Record, uint32)
	loadPosition    func(Object) types.Pointf
	loadRadius      func(Object) float32
	loadPrevious    func(Record) types.Pointf
	storePrevious   func(Record, types.Pointf)
	loadAnchor      func(Record) types.Pointf
	storeAnchor     func(Record, types.Pointf)
	loadLevel       func(Record) uint32
	randomInt       func(int, int) int
	spawnFlame      func(int, types.Pointf)
	loadTypeID      func(int) uint32
	storeTypeID     func(int, uint32)
	nameID          func(string) uint32
	allocateFlame   func(uint32) Object
	createFlame     func(Object, types.Pointf)
	audioFlame      func(types.Pointf)
	decayFlame      func(Object, uint32)
	tickRate        func() uint32
	roundTowardZero bool
}

//go:noinline
func firewalkSubtract53_52ED40(a, b float64, chop bool) float64 {
	if math.IsInf(a, 0) && a == b {
		// x87's masked invalid subtraction stores its negative indefinite
		// NaN; ARM64's default NaN has the opposite sign.
		return math.Float64frombits(0xfff8000000000000)
	}
	if chop && !math.IsNaN(a) && !math.IsNaN(b) {
		return monsterMoveToRunAddChop53_544434(a, -b)
	}
	return a - b
}

func firewalkSpill32_52ED40(value float64, chop bool) float32 {
	if chop {
		return float32(monsterMoveToRunSpill544440(value))
	}
	return float32(value)
}

//go:noinline
func firewalkDistance53_52ED40(position, previous types.Pointf, radius float32, chop bool) float64 {
	dx := firewalkSubtract53_52ED40(float64(position.X), float64(previous.X), chop)
	dy := firewalkSubtract53_52ED40(float64(position.Y), float64(previous.Y), chop)
	var distance float64
	if chop {
		square := monsterMoveToRunAddChop53_544434(monsterMoveToRunSquareChop53_544434(dy), monsterMoveToRunSquareChop53_544434(dx))
		distance = monsterMoveForceSqrtChop53_50D50C(square)
	} else {
		ySquare, xSquare := float64(dy*dy), float64(dx*dx)
		distance = math.Sqrt(float64(ySquare + xSquare))
	}
	return firewalkSubtract53_52ED40(distance, float64(radius), chop)
}

// spellFirewalkUpdate52ED40 follows GAME.EXE 0052ED40. The first tick records
// the target position; later ticks leave two flames after the target moves
// farther than its radius plus 15 units from that position.
func spellFirewalkUpdate52ED40[Record, Object comparable](record Record, h firewalkHooks52ED40[Record, Object]) int32 {
	var nilObject Object
	target := h.loadTarget(record)
	if target == nilObject || h.loadFlags(target)&0x8020 != 0 {
		return 1
	}
	if h.loadTypeID != nil && h.loadTypeID(0) == 0 {
		for i, name := range [3]string{"SmallFlame", "MediumFlame", "Flame"} {
			h.storeTypeID(i, h.nameID(name))
		}
	}
	if h.loadFrame60(record) == h.loadFrame64(record) {
		position := h.loadPosition(h.loadTarget(record))
		h.storePrevious(record, position)
		h.storeAnchor(record, position)
		h.storeFrame64(record, h.loadFrame64(record)+1)
		return 0
	}

	// Name lookup can change the record's target. The original reloads this
	// link, then keeps the level's signed comparison across both RNG calls.
	target = h.loadTarget(record)
	level := int32(h.loadLevel(record))
	if !(firewalkDistance53_52ED40(h.loadPosition(target), h.loadPrevious(record), h.loadRadius(target), h.roundTowardZero) > 15) {
		return 0
	}
	maximum := 0
	if level >= 2 {
		maximum = 1
		if level >= 4 {
			maximum = 2
		}
	}
	flamePos := h.loadPrevious(record)
	for range 2 {
		kind := h.randomInt(0, maximum)
		if h.allocateFlame != nil {
			flame := h.allocateFlame(h.loadTypeID(kind))
			if flame != nilObject {
				h.createFlame(flame, flamePos)
				h.audioFlame(flamePos)
				h.decayFlame(flame, 25*h.tickRate())
			}
		} else {
			h.spawnFlame(kind, flamePos)
		}
		anchor := h.loadAnchor(record)
		// Both differences are explicit binary32 spills before the ordered
		// nonzero tests. The midpoint then uses the spilled deltas, not a
		// second subtraction of the original coordinates.
		dx := firewalkSpill32_52ED40(firewalkSubtract53_52ED40(float64(flamePos.X), float64(anchor.X), h.roundTowardZero), h.roundTowardZero)
		dy := firewalkSpill32_52ED40(firewalkSubtract53_52ED40(float64(flamePos.Y), float64(anchor.Y), h.roundTowardZero), h.roundTowardZero)
		if math.Abs(float64(dx)) > 0 && math.Abs(float64(dy)) > 0 {
			flamePos.X = firewalkSpill32_52ED40(firewalkSubtract53_52ED40(float64(flamePos.X), float64(dx)*0.5, h.roundTowardZero), h.roundTowardZero)
			flamePos.Y = firewalkSpill32_52ED40(firewalkSubtract53_52ED40(float64(flamePos.Y), float64(dy)*0.5, h.roundTowardZero), h.roundTowardZero)
		}
	}
	h.storeAnchor(record, h.loadPrevious(record))
	h.storePrevious(record, h.loadPosition(h.loadTarget(record)))
	return 0
}

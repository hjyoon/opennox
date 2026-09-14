package server

import (
	"math"

	"github.com/opennox/libs/types"
)

// firewalkHooks52ED40 keeps object and duration identities native-width while
// preserving the original callback's frame and position state machine.
type firewalkHooks52ED40[Record, Object comparable] struct {
	loadTarget    func(Record) Object
	loadFlags     func(Object) uint32
	loadFrame60   func(Record) uint32
	loadFrame64   func(Record) uint32
	storeFrame64  func(Record, uint32)
	loadPosition  func(Object) types.Pointf
	loadRadius    func(Object) float32
	loadPrevious  func(Record) types.Pointf
	storePrevious func(Record, types.Pointf)
	loadAnchor    func(Record) types.Pointf
	storeAnchor   func(Record, types.Pointf)
	loadLevel     func(Record) uint32
	randomInt     func(int, int) int
	spawnFlame    func(int, types.Pointf)
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
	if h.loadFrame60(record) == h.loadFrame64(record) {
		position := h.loadPosition(h.loadTarget(record))
		h.storePrevious(record, position)
		h.storeAnchor(record, position)
		h.storeFrame64(record, h.loadFrame64(record)+1)
		return 0
	}

	previous := h.loadPrevious(record)
	position := h.loadPosition(target)
	dx := float64(position.X) - float64(previous.X)
	dy := float64(position.Y) - float64(previous.Y)
	if math.Sqrt(dx*dx+dy*dy)-float64(h.loadRadius(target)) <= 15 {
		return 0
	}
	maximum := 0
	if level := h.loadLevel(record); level >= 2 {
		maximum = 1
		if level >= 4 {
			maximum = 2
		}
	}
	flamePos := previous
	for range 2 {
		h.spawnFlame(h.randomInt(0, maximum), flamePos)
		anchor := h.loadAnchor(record)
		if flamePos.X-anchor.X != 0 && flamePos.Y-anchor.Y != 0 {
			flamePos.X -= (flamePos.X - anchor.X) * 0.5
			flamePos.Y -= (flamePos.Y - anchor.Y) * 0.5
		}
	}
	h.storeAnchor(record, h.loadPrevious(record))
	h.storePrevious(record, h.loadPosition(h.loadTarget(record)))
	return 0
}

package server

import "github.com/opennox/libs/types"

const (
	rewardAnkhMarkerTypeName4F2110     = "RewardMarker"
	rewardAnkhMarkerPlusTypeName4F2110 = "RewardMarkerPlus"
	rewardAnkhObjectTypeName4F2110     = "Ankh"
	rewardAnkhRandomPath4F2110         = `C:\NoxPost\src\server\GameMech\Reward.c`
	rewardAnkhRandomLine4F2110         = int32(2460)
	rewardAnkhActiveMask4F2110         = uint8(0x80)
)

type rewardAnkhReplaceHooks4F2110[O comparable, D any] struct {
	loadMarkerCache  func() uint32
	storeMarkerCache func(uint32)
	loadPlusCache    func() uint32
	storePlusCache   func(uint32)
	lookupType       func(string) uint32

	firstObject     func() O
	nextObject      func(O) O
	loadTypeInd     func(O) uint16
	loadInitData    func(O) D
	loadCategoryLow func(D) uint8
	randomInt       func(int32, int32, string, int32) int32
	newObject       func(string) O
	loadPosX        func(O) float32
	loadPosY        func(O) float32
	createAt        func(O, O, types.Pointf)
	delayedDelete   func(O)
}

// rewardAnkhReplace4F2110 reconstructs GAME.EXE 004F2110. It counts active
// RewardMarker and RewardMarkerPlus objects, draws one ordinal, then starts a
// fresh traversal and replaces that marker with an Ankh.
//
// The original calls its inclusive RNG even when the count is zero, producing
// the range 0..-1. It reloads the primary fixed-width type cache per element
// and reloads the plus cache only after a primary mismatch, reads the first
// InitData byte without a nil guard, and obtains the successor only after
// processing the current object. If creation at the selected ordinal returns
// nil, the ordinal is deliberately not advanced, so every later active marker
// retries the Ankh factory until one succeeds. Successful creation loads marker
// Y before X, creates the Ankh before deleting the marker, and returns without
// asking for the marker's successor.
func rewardAnkhReplace4F2110[O comparable, D any](hooks rewardAnkhReplaceHooks4F2110[O, D]) {
	var zero O
	var count int32
	markerType := hooks.loadMarkerCache()
	if markerType == 0 {
		markerType = hooks.lookupType(rewardAnkhMarkerTypeName4F2110)
		hooks.storeMarkerCache(markerType)
		plusType := hooks.lookupType(rewardAnkhMarkerPlusTypeName4F2110)
		hooks.storePlusCache(plusType)
	}

	for current := hooks.firstObject(); current != zero; current = hooks.nextObject(current) {
		markerType = hooks.loadMarkerCache()
		typeInd := uint32(hooks.loadTypeInd(current))
		isMarker := typeInd == markerType
		if !isMarker {
			isMarker = typeInd == hooks.loadPlusCache()
		}
		if isMarker {
			data := hooks.loadInitData(current)
			if hooks.loadCategoryLow(data)&rewardAnkhActiveMask4F2110 != 0 {
				count++
			}
		}
	}

	selected := hooks.randomInt(0, count-1, rewardAnkhRandomPath4F2110, rewardAnkhRandomLine4F2110)
	var ordinal int32
	for current := hooks.firstObject(); current != zero; current = hooks.nextObject(current) {
		markerType = hooks.loadMarkerCache()
		typeInd := uint32(hooks.loadTypeInd(current))
		isMarker := typeInd == markerType
		if !isMarker {
			isMarker = typeInd == hooks.loadPlusCache()
		}
		if isMarker {
			data := hooks.loadInitData(current)
			if hooks.loadCategoryLow(data)&rewardAnkhActiveMask4F2110 != 0 {
				if ordinal == selected {
					ankh := hooks.newObject(rewardAnkhObjectTypeName4F2110)
					if ankh != zero {
						y := hooks.loadPosY(current)
						x := hooks.loadPosX(current)
						hooks.createAt(ankh, zero, types.Pointf{X: x, Y: y})
						hooks.delayedDelete(current)
						return
					}
				} else {
					ordinal++
				}
			}
		}
	}
}

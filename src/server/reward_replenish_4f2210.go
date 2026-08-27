package server

const (
	rewardReplenishMarkerTypeName4F2210     = "RewardMarker"
	rewardReplenishMarkerPlusTypeName4F2210 = "RewardMarkerPlus"
	rewardReplenishPotionTypeName4F2210     = "RedPotion"
	rewardReplenishRandomPath4F2210         = `C:\NoxPost\src\server\GameMech\Reward.c`
	rewardReplenishMarkerRandomLine4F2210   = int32(2631)
	rewardReplenishPotionRandomLine4F2210   = int32(2660)
	rewardReplenishFixedMarkerMask4F2210    = uint8(0x01)
	rewardReplenishActiveMask4F2210         = uint8(0x80)
)

type rewardReplenishHooks4F2210[O comparable, D any, A any] struct {
	loadQuestStage  func() uint32
	loadPlayerCount func() int32

	loadMarkerCache  func() uint32
	storeMarkerCache func(uint32)
	loadPlusCache    func() uint32
	storePlusCache   func(uint32)
	loadPotionCache  func() uint32
	storePotionCache func(uint32)
	lookupType       func(string) uint32

	firstObject   func() O
	nextObject    func(O) O
	loadTypeInd   func(O) uint16
	loadInitData  func(O) D
	loadFieldLow  func(D) uint8
	storeFieldLow func(D, uint8)

	allocObjects func(int32) A
	storeObject  func(A, int32, O)
	loadObject   func(A, int32) O
	freeObjects  func(A)

	randomInt     func(int32, int32, string, int32) int32
	delayedDelete func(O)
}

func rewardReplenishFraction4F2210(questStage uint32, playerCount int32) float32 {
	if questStage == 1 {
		return 0.5
	}
	switch playerCount {
	case 1, 2:
		return 0.4
	case 3, 4:
		return 0.7
	case 5, 6:
		return 1.0
	default:
		return 0
	}
}

// rewardReplenishRounded4F2210 models the original x87 expression followed by
// truncation toward zero. Object-list counts are non-negative and small enough
// that the conversion is representable in int32.
func rewardReplenishRounded4F2210(count int32, fraction float32) int32 {
	return int32(float64(count)*float64(fraction) + 0.5)
}

// rewardReplenishShuffle4F2210 is the original descending Fisher-Yates loop.
// The array access order is selected slot, tail slot, selected store, tail
// store. An out-of-range RNG result therefore faults at the same first access.
func rewardReplenishShuffle4F2210[O comparable, A any](
	objects A,
	count int32,
	path string,
	line int32,
	load func(A, int32) O,
	store func(A, int32, O),
	randomInt func(int32, int32, string, int32) int32,
) {
	for tail := count - 1; tail > 0; tail-- {
		selectedIndex := randomInt(0, tail, path, line)
		selected := load(objects, selectedIndex)
		tailObject := load(objects, tail)
		store(objects, selectedIndex, tailObject)
		store(objects, tail, selected)
	}
}

// rewardReplenish4F2210 reconstructs GAME.EXE 004F2210. It first counts
// inactive RewardMarker objects and RedPotion objects, allocates exact PE32
// pointer arrays from those first-pass counts, then starts a fresh traversal.
// Fixed RewardMarker and every RewardMarkerPlus become active immediately;
// other markers and potions are collected, shuffled independently, and scaled
// by quest stage or player count. Excess potions are delayed-deleted.
//
// Observable original quirks are intentional: only a zero primary cache
// initializes all three caches; type caches are reloaded live per comparison;
// InitData is dereferenced without a nil guard; successors are fetched after
// processing; second-pass growth can write beyond the first-pass allocation;
// and an allocated array leaks when the second pass is empty or collects none.
func rewardReplenish4F2210[O comparable, D any, A any](
	hooks rewardReplenishHooks4F2210[O, D, A],
) {
	var zero O
	var markerObjects A
	var potionObjects A

	questStage := hooks.loadQuestStage()
	playerCount := hooks.loadPlayerCount()
	if hooks.loadMarkerCache() == 0 {
		hooks.storeMarkerCache(hooks.lookupType(rewardReplenishMarkerTypeName4F2210))
		hooks.storePlusCache(hooks.lookupType(rewardReplenishMarkerPlusTypeName4F2210))
		hooks.storePotionCache(hooks.lookupType(rewardReplenishPotionTypeName4F2210))
	}
	fraction := rewardReplenishFraction4F2210(questStage, playerCount)

	var markerCount int32
	var potionCount int32
	current := hooks.firstObject()
	if current == zero {
		return
	}
	for current != zero {
		markerType := hooks.loadMarkerCache()
		typeInd := uint32(hooks.loadTypeInd(current))
		if typeInd == markerType {
			data := hooks.loadInitData(current)
			if hooks.loadFieldLow(data)&rewardReplenishFixedMarkerMask4F2210 == 0 {
				markerCount++
			}
		} else if typeInd == hooks.loadPotionCache() {
			potionCount++
		}
		current = hooks.nextObject(current)
	}

	if markerCount != 0 {
		markerObjects = hooks.allocObjects(markerCount)
		if potionCount == 0 {
			goto secondPass
		}
	} else if potionCount == 0 {
		return
	}
	potionObjects = hooks.allocObjects(potionCount)

secondPass:
	var collectedMarkers int32
	var collectedPotions int32
	current = hooks.firstObject()
	if current == zero {
		return
	}
	for current != zero {
		markerType := hooks.loadMarkerCache()
		typeInd := uint32(hooks.loadTypeInd(current))
		if typeInd == markerType {
			data := hooks.loadInitData(current)
			value := hooks.loadFieldLow(data)
			if value&rewardReplenishFixedMarkerMask4F2210 != 0 {
				hooks.storeFieldLow(data, value|rewardReplenishActiveMask4F2210)
			} else {
				hooks.storeObject(markerObjects, collectedMarkers, current)
				collectedMarkers++
			}
		} else if typeInd == hooks.loadPlusCache() {
			data := hooks.loadInitData(current)
			value := hooks.loadFieldLow(data)
			hooks.storeFieldLow(data, value|rewardReplenishActiveMask4F2210)
		} else if typeInd == hooks.loadPotionCache() {
			hooks.storeObject(potionObjects, collectedPotions, current)
			collectedPotions++
		}
		current = hooks.nextObject(current)
	}

	if collectedMarkers != 0 {
		desired := rewardReplenishRounded4F2210(collectedMarkers, fraction)
		rewardReplenishShuffle4F2210(
			markerObjects,
			collectedMarkers,
			rewardReplenishRandomPath4F2210,
			rewardReplenishMarkerRandomLine4F2210,
			hooks.loadObject,
			hooks.storeObject,
			hooks.randomInt,
		)
		for index := int32(0); index < desired; index++ {
			object := hooks.loadObject(markerObjects, index)
			data := hooks.loadInitData(object)
			value := hooks.loadFieldLow(data)
			hooks.storeFieldLow(data, value|rewardReplenishActiveMask4F2210)
		}
		hooks.freeObjects(markerObjects)
	}

	if collectedPotions != 0 {
		desired := rewardReplenishRounded4F2210(collectedPotions, fraction)
		rewardReplenishShuffle4F2210(
			potionObjects,
			collectedPotions,
			rewardReplenishRandomPath4F2210,
			rewardReplenishPotionRandomLine4F2210,
			hooks.loadObject,
			hooks.storeObject,
			hooks.randomInt,
		)
		for index := desired; index < collectedPotions; index++ {
			hooks.delayedDelete(hooks.loadObject(potionObjects, index))
		}
		hooks.freeObjects(potionObjects)
	}
}

package server

import (
	"unsafe"

	"github.com/opennox/libs/types"
)

const (
	rewardContainerMarkerTypeName4F1F20     = "RewardMarker"
	rewardContainerMarkerPlusTypeName4F1F20 = "RewardMarkerPlus"
	rewardContainerQuiverTypeName4F1F20     = "Quiver"

	rewardContainerActiveMask4F1F20      = uint8(0x80)
	rewardContainerWeaponClass4F1F20     = uint32(0x01000000)
	rewardContainerBowSubclassMask4F1F20 = uint8(0x0c)
	rewardContainerQuiverRadius4F1F20    = float32(30.0)
)

type rewardContainerHooks4F1F20[O comparable, D any] struct {
	loadQuestStage func() uint32

	loadMarkerCache  func() uint32
	storeMarkerCache func(uint32)
	loadPlusCache    func() uint32
	storePlusCache   func(uint32)
	lookupType       func(string) uint32

	preprocessMarkers func()
	preprocessRewards func()

	firstObject func() O
	nextObject  func(O) O
	loadTypeInd func(O) uint16
	loadInit    func(O) unsafe.Pointer
	isChestInit func(unsafe.Pointer) bool

	loadInitData      func(O) D
	loadField216Low   func(D) uint8
	firstInventory    func(O) O
	nextInventoryItem func(O) O

	activateMarker  func(O, uint32) O
	detachInventory func(O, O)
	delayedDelete   func(O)
	inventoryPut    func(O, O, uint32)

	loadClass       func(O) uint32
	loadSubclassLow func(O) uint8
	newObject       func(string) O
	loadPosX        func(O) float32
	loadPosY        func(O) float32
	createAt        func(O, O, types.Pointf)
	randomReachable func(float32, O, *types.Pointf)
}

// rewardContainerProcess4F1F20 reconstructs GAME.EXE 004F1F20. The two
// reward-container type caches remain distinct from 004F0720's activation
// cache. Both object loops cache their next pointer before reading any other
// field or invoking a callback, and they reload the live type caches for every
// element. Those details allow activation, inventory, and delete callbacks to
// mutate the traversed lists and cache storage without changing the original
// continuation point.
//
// World markers test the low byte of Field216, create rewards at the marker's
// position, optionally place a Quiver near bow/crossbow rewards, and are always
// delayed-deleted. Chest markers activate at stage+1 with uint32 wrapping,
// detach and delete before inserting any result, and optionally insert a
// Quiver. There are deliberately no nil-data or callback guards.
func rewardContainerProcess4F1F20[O comparable, D any](
	hooks rewardContainerHooks4F1F20[O, D],
) {
	var zero O
	stage := hooks.loadQuestStage()
	markerType := hooks.loadMarkerCache()
	if markerType == 0 {
		markerType = hooks.lookupType(rewardContainerMarkerTypeName4F1F20)
		hooks.storeMarkerCache(markerType)
		plusType := hooks.lookupType(rewardContainerMarkerPlusTypeName4F1F20)
		hooks.storePlusCache(plusType)
	}

	hooks.preprocessMarkers()
	hooks.preprocessRewards()

	for current := hooks.firstObject(); current != zero; {
		next := hooks.nextObject(current)
		markerType = hooks.loadMarkerCache()
		typeInd := uint32(hooks.loadTypeInd(current))
		isMarker := typeInd == markerType
		if !isMarker {
			isMarker = typeInd == hooks.loadPlusCache()
		}
		if isMarker {
			data := hooks.loadInitData(current)
			if hooks.loadField216Low(data)&rewardContainerActiveMask4F1F20 != 0 {
				reward := hooks.activateMarker(current, stage)
				if reward != zero {
					// x86 pushed Y before X for the four-argument CreateAt call.
					y := hooks.loadPosY(current)
					x := hooks.loadPosX(current)
					hooks.createAt(reward, zero, types.Pointf{X: x, Y: y})
					if hooks.loadClass(reward)&rewardContainerWeaponClass4F1F20 != 0 &&
						hooks.loadSubclassLow(reward)&rewardContainerBowSubclassMask4F1F20 != 0 {
						quiver := hooks.newObject(rewardContainerQuiverTypeName4F1F20)
						if quiver != zero {
							point := types.Pointf{
								X: hooks.loadPosX(reward),
								Y: hooks.loadPosY(reward),
							}
							hooks.randomReachable(rewardContainerQuiverRadius4F1F20, reward, &point)
							hooks.createAt(quiver, zero, point)
						}
					}
				}
			}
			hooks.delayedDelete(current)
		} else {
			init := hooks.loadInit(current)
			if hooks.isChestInit(init) {
				for item := hooks.firstInventory(current); item != zero; {
					nextItem := hooks.nextInventoryItem(item)
					markerType = hooks.loadMarkerCache()
					itemType := uint32(hooks.loadTypeInd(item))
					isMarker = itemType == markerType
					if !isMarker {
						isMarker = itemType == hooks.loadPlusCache()
					}
					if isMarker {
						reward := hooks.activateMarker(item, stage+1)
						hooks.detachInventory(current, item)
						hooks.delayedDelete(item)
						if reward != zero {
							hooks.inventoryPut(current, reward, 0)
							if hooks.loadClass(reward)&rewardContainerWeaponClass4F1F20 != 0 &&
								hooks.loadSubclassLow(reward)&rewardContainerBowSubclassMask4F1F20 != 0 {
								quiver := hooks.newObject(rewardContainerQuiverTypeName4F1F20)
								if quiver != zero {
									hooks.inventoryPut(current, quiver, 0)
								}
							}
						}
					}
					item = nextItem
				}
			}
		}
		current = next
	}
}

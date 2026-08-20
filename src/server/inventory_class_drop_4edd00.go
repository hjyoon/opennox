package server

const (
	respawnInventoryClassRadius4EDD00  = float32(60)
	dropPlayerInventoryClassMask4EDD70 = uint32(0x10000000)
	dropPlayerInventoryRadius4EDD70    = float32(50)
)

// respawnInventoryClassHooks4EDD00 exposes every observable read and callback
// boundary in GAME.EXE 004EDD00. The class mask is loaded only after a
// non-empty inventory is observed. Point Y is read before X after the random
// helper, matching the original IA-32 argument construction.
type respawnInventoryClassHooks4EDD00[O, I comparable, P any] struct {
	loadOwnerArg      func() O
	loadInventoryHead func(O) I
	loadClassMaskArg  func() uint32
	loadInventoryNext func(I) I
	loadItemClass     func(I) uint32
	detachInventory   func(O, I)
	ownerPosition     func(O) *P
	randomReachable   func(radius float32, center, output *P) *P
	loadPointY        func(*P) float32
	loadPointX        func(*P) float32
	createAt          func(item I, owner O, x, y float32)
}

// respawnInventoryClass4EDD00 preserves GAME.EXE 004EDD00. Each successor is
// cached before the item's full class dword is read and before detach, random,
// or create side effects. A single stack-local point is reused for every
// matching item, the random helper's return pointer is ignored, and createAt
// receives the zero owner value.
func respawnInventoryClass4EDD00[O, I comparable, P any](
	hooks respawnInventoryClassHooks4EDD00[O, I, P],
) {
	var zeroItem I
	var zeroOwner O

	owner := hooks.loadOwnerArg()
	current := hooks.loadInventoryHead(owner)
	if current == zeroItem {
		return
	}
	classMask := hooks.loadClassMaskArg()
	var point P
	for {
		next := hooks.loadInventoryNext(current)
		itemClass := hooks.loadItemClass(current)
		if itemClass&classMask != 0 {
			hooks.detachInventory(owner, current)
			_ = hooks.randomReachable(
				respawnInventoryClassRadius4EDD00,
				hooks.ownerPosition(owner),
				&point,
			)
			y := hooks.loadPointY(&point)
			x := hooks.loadPointX(&point)
			hooks.createAt(current, zeroOwner, x, y)
		}
		if next == zeroItem {
			return
		}
		current = next
	}
}

// dropPlayerInventoryClassHooks4EDD70 exposes GAME.EXE 004EDD70 without
// assuming the width or layout of Object. The first inventory pointer and the
// next-player pointer remain live reads at their original loop boundaries.
type dropPlayerInventoryClassHooks4EDD70[O, I comparable, P any] struct {
	firstPlayer       func() O
	loadInventoryHead func(O) I
	loadItemClass     func(I) uint32
	loadInventoryNext func(I) I
	playerPosition    func(O) *P
	randomReachable   func(radius float32, center, output *P) *P
	drop              func(O, I, *P) int32
	nextPlayer        func(O) O
}

// dropPlayerInventoryClass4EDD70 preserves GAME.EXE 004EDD70. Item class is
// cached before its successor, and that successor is cached before the class
// gate and callbacks. One stack-local point is reused across every player and
// item. Random-helper and drop results are both discarded. The next player is
// fetched only after the current player's complete inventory scan.
func dropPlayerInventoryClass4EDD70[O, I comparable, P any](
	hooks dropPlayerInventoryClassHooks4EDD70[O, I, P],
) {
	var zeroObject O
	var zeroItem I
	var point P

	for player := hooks.firstPlayer(); player != zeroObject; player = hooks.nextPlayer(player) {
		for item := hooks.loadInventoryHead(player); item != zeroItem; {
			itemClass := hooks.loadItemClass(item)
			next := hooks.loadInventoryNext(item)
			if itemClass&dropPlayerInventoryClassMask4EDD70 != 0 {
				_ = hooks.randomReachable(
					dropPlayerInventoryRadius4EDD70,
					hooks.playerPosition(player),
					&point,
				)
				_ = hooks.drop(player, item, &point)
			}
			item = next
		}
	}
}

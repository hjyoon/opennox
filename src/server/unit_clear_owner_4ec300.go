package server

const (
	unitClearOwnerPlayerClass4EC300   = uint32(0x04)
	unitClearOwnerMonsterClass4EC300  = uint32(0x02)
	unitClearOwnerUnitClassMask4EC300 = uint32(0x06)
	unitClearOwnerMonitorBit4EC300    = uint32(0x80)
)

type unitClearOwnerHooks4EC300[O comparable, U, P any] struct {
	loadOwner       func(O) O
	loadClass       func(O) uint32
	isMonitored     func(O, O) bool
	loadSubClass    func(O) uint32
	loadPlayerData  func(O) U
	storeSubClass   func(O, uint32)
	loadPlayer      func(U) P
	loadPlayerIndex func(P) uint8
	netFxShield     func(uint8, O)
	unmarkMinimap   func(uint8, O, uint32)
	loadFirstOwned  func(O) O
	loadNextOwned   func(O) O
	storeNextOwned  func(O, O)
	storeFirstOwned func(O, O)
	storeOwner      func(O, O)
	resetMonster    func(O)
	markUnitUpdate  func(O)
}

// unitClearOwner4EC300 preserves GAME.EXE 004EC300. The entry owner is used
// only for the Player/monitored gate. Every later owner read is live, including
// the owner used for notification data and the owner whose owned list is
// repaired. A monitored-object notification also reloads the Player pointer
// between its two network callbacks. The object keeps its next-owned link.
func unitClearOwner4EC300[O comparable, U, P any](obj O, hooks unitClearOwnerHooks4EC300[O, U, P]) {
	var zero O
	if obj == zero {
		return
	}
	owner := hooks.loadOwner(obj)
	if owner == zero {
		return
	}

	if hooks.loadClass(owner)&unitClearOwnerPlayerClass4EC300 != 0 && hooks.isMonitored(owner, obj) {
		liveOwner := hooks.loadOwner(obj)
		subClass := hooks.loadSubClass(obj) &^ unitClearOwnerMonitorBit4EC300
		data := hooks.loadPlayerData(liveOwner)
		hooks.storeSubClass(obj, subClass)

		player := hooks.loadPlayer(data)
		ind := hooks.loadPlayerIndex(player)
		hooks.netFxShield(ind, obj)

		player = hooks.loadPlayer(data)
		ind = hooks.loadPlayerIndex(player)
		hooks.unmarkMinimap(ind, obj, 1)
	}

	owner = hooks.loadOwner(obj)
	cur := hooks.loadFirstOwned(owner)
	var prev O
	for cur != zero {
		if cur == obj {
			break
		}
		prev = cur
		cur = hooks.loadNextOwned(cur)
	}
	next := hooks.loadNextOwned(obj)
	if prev != zero {
		hooks.storeNextOwned(prev, next)
	} else {
		hooks.storeFirstOwned(owner, next)
	}

	class := hooks.loadClass(obj)
	hooks.storeOwner(obj, zero)
	if class&unitClearOwnerMonsterClass4EC300 != 0 {
		hooks.resetMonster(obj)
	}
	if hooks.loadClass(obj)&unitClearOwnerUnitClassMask4EC300 != 0 {
		hooks.markUnitUpdate(obj)
	}
}

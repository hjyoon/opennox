package server

const (
	flagPickupCTFHomeLimit4EA490      = float64(5)
	flagPickupCTFFlagClass4EA490      = uint32(0x10000000)
	flagPickupCTFScoreMode4EA490      = uint32(32)
	flagPickupCTFScoreDelta4EA490     = int32(1)
	flagPickupCTFWinnerGameFlag4EA490 = uint32(8)
	flagPickupCTFReturnMessage4EA490  = uint32(4)
	flagPickupCTFScoreMessage4EA490   = uint32(5)
	flagPickupCTFPickupMessage4EA490  = uint32(6)
	flagPickupCTFRecipient4EA490      = uint32(255)
	flagPickupCTFMinimapFlag4EA490    = uint32(1)
	flagPickupCTFWeaponFlag4EA490     = uint32(1)
	flagPickupCTFNoTeamMessage4EA490  = "objcoll.c:FlagNoTeam"
)

type flagPickupCTFHooks4EA490[O, U, T, P comparable] struct {
	loadUpdate       func(O) U
	flagIndex        func(O) uint32
	loadTeamID       func(O) uint8
	teamsSame        func(O, O) int32
	loadPosX         func(O) float32
	loadPosY         func(O) float32
	loadHomeX        func(U) float32
	loadHomeY        func(U) float32
	moveHome         func(O, U)
	loadNetCode      func(O) uint32
	informReturn     func(uint32)
	informFlag       func(uint32, uint32, uint32)
	storeFlagState   func(U, uint32)
	flagStatus       func(uint8, uint8, uint8, uint16) int32
	firstInventory   func(O) O
	nextInventory    func(O) O
	loadClass        func(O) uint32
	gameData         func(uint32) uint16
	changeScore      func(O, int32)
	reportLesson     func(O)
	hasTeam          func(O) int32
	teamByID         func(uint8) T
	loadTeamScore    func(T) int32
	changeTeamScore  func(T, int32)
	observerMode     func() uint32
	playerFromUpdate func(U) P
	observerUpdate   func(P, P)
	detachInventory  func(O, O)
	createAt         func(O, O, float32, float32)
	raise            func(O, float32)
	markMinimap      func(O, uint32)
	firstTeam        func() T
	nextTeam         func(T) T
	setGameFlags     func(uint32)
	flagWinner       func(T, uint32)
	inventoryHolder  func(O) O
	teamEligible     func(T) int32
	forceDrop        func(O, O)
	finalizeDelete   func(O)
	inventoryPut     func(O, O, int32)
	markPlayerPickup func(U, uint32)
	reportObject     func(uint32, O)
	unmarkMinimap    func(O, uint32)
	purgeBuffs       func(O)
	priorityMessage  func(O, string, uint32)
}

// flagPickupCTFOutsideHome4EA490 matches the two x87 FABS/FCOMP gates. A
// binary64 subtraction is exact for two binary32 inputs, and an unordered
// comparison (NaN) remains false just as in the original status-word test.
func flagPickupCTFOutsideHome4EA490(pos, home float32) bool {
	delta := float64(pos) - float64(home)
	if delta < 0 {
		delta = -delta
	}
	return delta > flagPickupCTFHomeLimit4EA490
}

// flagPickupCTF4EA490 preserves GAME.EXE 004EA490. The collision argument is
// present in the three-argument callback ABI but is not inspected. Update-data
// pointers, flag indexes, and selected team bytes are cached only where the
// original kept them in registers or stack locals; other fields remain live
// across effect callbacks.
func flagPickupCTF4EA490[O, U, T, P comparable, C any](
	source, target O,
	_ C,
	hooks flagPickupCTFHooks4EA490[O, U, T, P],
) {
	targetUpdate := hooks.loadUpdate(target)
	sourceIndex := hooks.flagIndex(source)
	sourceTeamID := hooks.loadTeamID(source)
	if hooks.teamsSame(target, source) != 0 {
		flagPickupCTFSameTeam4EA490(source, target, targetUpdate, sourceIndex, hooks)
		return
	}
	flagPickupCTFEnemyTeam4EA490(source, target, targetUpdate, sourceIndex, sourceTeamID, hooks)
}

func flagPickupCTFSameTeam4EA490[O, U, T, P comparable](
	source, target O,
	targetUpdate U,
	sourceIndex uint32,
	hooks flagPickupCTFHooks4EA490[O, U, T, P],
) {
	sourceUpdate := hooks.loadUpdate(source)
	x := hooks.loadPosX(source)
	homeX := hooks.loadHomeX(sourceUpdate)
	if flagPickupCTFOutsideHome4EA490(x, homeX) {
		flagPickupCTFReturnHome4EA490(source, target, sourceUpdate, sourceIndex, hooks)
		return
	}
	y := hooks.loadPosY(source)
	homeY := hooks.loadHomeY(sourceUpdate)
	if flagPickupCTFOutsideHome4EA490(y, homeY) {
		flagPickupCTFReturnHome4EA490(source, target, sourceUpdate, sourceIndex, hooks)
		return
	}

	var zeroObject O
	for item := hooks.firstInventory(target); item != zeroObject; item = hooks.nextInventory(item) {
		if hooks.loadClass(item)&flagPickupCTFFlagClass4EA490 == 0 {
			continue
		}
		threshold := hooks.gameData(flagPickupCTFScoreMode4EA490)
		itemUpdate := hooks.loadUpdate(item)
		itemIndex := hooks.flagIndex(item)
		itemTeamID := hooks.loadTeamID(item)
		hooks.changeScore(target, flagPickupCTFScoreDelta4EA490)
		hooks.reportLesson(target)
		if hooks.hasTeam(target) != 0 {
			teamID := hooks.loadTeamID(target)
			team := hooks.teamByID(teamID)
			score := hooks.loadTeamScore(team)
			hooks.changeTeamScore(team, score+flagPickupCTFScoreDelta4EA490)
			if hooks.observerMode() != 0 {
				var zeroUpdate U
				if targetUpdate != zeroUpdate {
					player := hooks.playerFromUpdate(targetUpdate)
					var zeroPlayer P
					hooks.observerUpdate(player, zeroPlayer)
				}
			}
		}
		hooks.detachInventory(target, item)
		itemHomeY := hooks.loadHomeY(itemUpdate)
		itemHomeX := hooks.loadHomeX(itemUpdate)
		hooks.createAt(item, zeroObject, itemHomeX, itemHomeY)
		hooks.raise(item, 0)
		hooks.markMinimap(item, flagPickupCTFMinimapFlag4EA490)
		hooks.storeFlagState(itemUpdate, 0)
		hooks.flagStatus(itemTeamID, 0, uint8(itemIndex), 0)
		netCode := hooks.loadNetCode(target)
		hooks.informFlag(flagPickupCTFScoreMessage4EA490, netCode, itemIndex)

		if int32(threshold) > 0 {
			var zeroTeam T
			for team := hooks.firstTeam(); team != zeroTeam; team = hooks.nextTeam(team) {
				if hooks.loadTeamScore(team) >= int32(threshold) {
					hooks.setGameFlags(flagPickupCTFWinnerGameFlag4EA490)
					hooks.flagWinner(team, 0)
					return
				}
			}
			return
		}
	}
}

func flagPickupCTFReturnHome4EA490[O, U, T, P comparable](
	source, target O,
	sourceUpdate U,
	sourceIndex uint32,
	hooks flagPickupCTFHooks4EA490[O, U, T, P],
) {
	netCode := hooks.loadNetCode(target)
	hooks.moveHome(source, sourceUpdate)
	hooks.informReturn(netCode)
	hooks.storeFlagState(sourceUpdate, 0)
	targetTeamID := hooks.loadTeamID(target)
	hooks.flagStatus(targetTeamID, 0, uint8(sourceIndex), 0)
}

func flagPickupCTFEnemyTeam4EA490[O, U, T, P comparable](
	source, target O,
	targetUpdate U,
	sourceIndex uint32,
	sourceTeamID uint8,
	hooks flagPickupCTFHooks4EA490[O, U, T, P],
) {
	sourceUpdate := hooks.loadUpdate(source)
	var zeroObject O
	if hooks.inventoryHolder(source) != zeroObject {
		return
	}
	liveTeamID := hooks.loadTeamID(source)
	team := hooks.teamByID(liveTeamID)
	var zeroTeam T
	if team == zeroTeam || hooks.teamEligible(team) == 0 {
		hooks.priorityMessage(target, flagPickupCTFNoTeamMessage4EA490, 0)
		return
	}

	for item := hooks.firstInventory(target); item != zeroObject; item = hooks.nextInventory(item) {
		if hooks.loadClass(item)&flagPickupCTFFlagClass4EA490 != 0 {
			hooks.forceDrop(target, item)
			break
		}
	}
	hooks.finalizeDelete(source)
	pickupIndex := hooks.flagIndex(source)
	hooks.inventoryPut(target, source, 1)
	hooks.markPlayerPickup(targetUpdate, flagPickupCTFWeaponFlag4EA490)
	hooks.reportObject(flagPickupCTFRecipient4EA490, source)
	netCode := hooks.loadNetCode(target)
	hooks.informFlag(flagPickupCTFPickupMessage4EA490, netCode, pickupIndex)
	hooks.unmarkMinimap(source, flagPickupCTFMinimapFlag4EA490)
	hooks.storeFlagState(sourceUpdate, 0)
	carrierNetCode := uint16(hooks.loadNetCode(target))
	hooks.flagStatus(sourceTeamID, 1, uint8(sourceIndex), carrierNetCode)
	hooks.purgeBuffs(target)
}

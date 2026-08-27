package server

const (
	pickupTreasurePlayerClassLow4F3580 = uint8(0x04)
	pickupTreasureGameMode4F3580       = uint32(0x40)
	pickupTreasureCompleteFlag4F3580   = uint32(0x08)
	pickupTreasureAudio4F3580          = uint32(307)
)

// pickupTreasureHooks4F3580 exposes every callback and delayed field load in
// GAME.EXE 004F3580. UpdateData is cached before the Scavenger-mode gate, but
// its Player pointer is deliberately reloaded after callbacks. Team and player
// fields that x86 reads inside loops remain live through separate hooks.
type pickupTreasureHooks4F3580[O comparable, U, P any, T comparable] struct {
	loadArg4      func() int32
	loadArg3      func() int32
	defaultPickup func(O, O, int32, int32) int32

	loadClassLow func(O) uint8
	loadUpdate   func(O) U
	gameFlag     func(uint32) int32
	audio        func(uint32, O, int32, uint32)
	loadPlayer   func(U) P
	loadCount    func(P) uint32
	storeCount   func(P, uint32)
	treasureMax  func() uint32
	storeMax     func(P, uint32)
	report       func(O)

	hasTeam         func(O) int32
	loadObjectTeam  func(O) uint8
	findTeam        func(uint8) T
	loadTeamID      func(T) uint8
	teamContains    func(O, uint8) int32
	firstPlayer     func() O
	nextPlayer      func(O) O
	setGameFlags    func(uint32)
	changeScore     func(O, int32)
	reportLesson    func(O)
	incrementDeaths func(O)
}

// pickupTreasure4F3580 preserves GAME.EXE 004F3580. DefaultPickup receives all
// four registered callback arguments with arg4 loaded before arg3. A nonzero
// DefaultPickup result is canonicalized to one; only Players in Scavenger mode
// update the wrapping treasure counters. Solo completion awards one lesson and
// eliminates every other player, while team completion sums only live members
// of the resolved team. The original has no nil guards.
func pickupTreasure4F3580[O comparable, U, P any, T comparable](
	owner, item O,
	hooks pickupTreasureHooks4F3580[O, U, P, T],
) int32 {
	arg4 := hooks.loadArg4()
	arg3 := hooks.loadArg3()
	if hooks.defaultPickup(owner, item, arg3, arg4) == 0 {
		return 0
	}
	if hooks.loadClassLow(owner)&pickupTreasurePlayerClassLow4F3580 == 0 {
		return 1
	}

	update := hooks.loadUpdate(owner)
	if hooks.gameFlag(pickupTreasureGameMode4F3580) == 0 {
		return 1
	}

	hooks.audio(pickupTreasureAudio4F3580, owner, 0, 0)
	player := hooks.loadPlayer(update)
	count := hooks.loadCount(player)
	hooks.storeCount(player, count+1)
	maximum := hooks.treasureMax()
	player = hooks.loadPlayer(update)
	hooks.storeMax(player, maximum)
	hooks.report(owner)

	if hooks.hasTeam(owner) == 0 {
		maximum = hooks.treasureMax()
		player = hooks.loadPlayer(update)
		if hooks.loadCount(player) == maximum {
			hooks.setGameFlags(pickupTreasureCompleteFlag4F3580)
			hooks.changeScore(owner, 1)
			hooks.reportLesson(owner)
			var nilObject O
			for current := hooks.firstPlayer(); current != nilObject; current = hooks.nextPlayer(current) {
				if current != owner {
					hooks.incrementDeaths(current)
					hooks.reportLesson(current)
				}
			}
		}
		return 1
	}

	teamID := hooks.loadObjectTeam(owner)
	team := hooks.findTeam(teamID)
	var nilTeam T
	if team == nilTeam {
		return 1
	}

	var total uint32
	var nilObject O
	for current := hooks.firstPlayer(); current != nilObject; current = hooks.nextPlayer(current) {
		teamID = hooks.loadTeamID(team)
		if hooks.teamContains(current, teamID) != 0 {
			currentUpdate := hooks.loadUpdate(current)
			currentPlayer := hooks.loadPlayer(currentUpdate)
			total += hooks.loadCount(currentPlayer)
		}
	}
	maximum = hooks.treasureMax()
	if total == maximum {
		hooks.setGameFlags(pickupTreasureCompleteFlag4F3580)
	}
	return 1
}

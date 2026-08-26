package server

type teamCreateAtHooks4191D0[V, O, T, P comparable] struct {
	loadGameBallType  func() uint16
	findTeam          func(uint8) T
	containsTeam      func(V, uint8) bool
	createTeam        func(uint8) T
	linkTeam          func(V, T) bool
	loadTeamID        func(T) uint8
	clientNetCode     func() uint32
	selectLocalTeam   func(uint8)
	afterAttach       func(T, V, int32, uint32, int32, uint16)
	commitMemberCount func(T)
	firstPlayerUnit   func() O
	nextPlayerUnit    func(O) O
	loadNetCode       func(O) uint32
	loadPlayer        func(O) P
	loadPlayerIndex   func(P) uint8
	resetPlayer       func(uint8)
	rebuildMasks      func(uint8)
	markUpdate        func(O)
	firstOwned        func(O) O
	nextOwned         func(O) O
	loadClassLow      func(O) uint8
}

// teamCreateAt4191D0 preserves GAME.EXE 004191D0's team-link and player-mark
// ordering. The original PE32 routine cached GameBall before its null guard,
// linked the object before notifications, published the member count after
// notifications, then scanned player units by live net code. A match resets
// that player's network relation state and marks the player plus class 0x02/
// 0x04 owned objects. Pointer traversal is supplied by typed hooks so no
// native address is narrowed to the original 32-bit object slots.
func teamCreateAt4191D0[V, O, T, P comparable](
	teamID uint8,
	value V,
	active int32,
	netCode uint32,
	flags int32,
	hooks teamCreateAtHooks4191D0[V, O, T, P],
) {
	gameBallType := hooks.loadGameBallType()
	var zeroValue V
	var zeroObject O
	var zeroTeam T
	var zeroPlayer P
	if value == zeroValue {
		return
	}

	team := hooks.findTeam(teamID)
	if team != zeroTeam {
		if hooks.containsTeam(value, teamID) {
			return
		}
	} else {
		team = hooks.createTeam(teamID)
	}
	if team == zeroTeam || !hooks.linkTeam(value, team) {
		return
	}

	liveTeamID := hooks.loadTeamID(team)
	if netCode == hooks.clientNetCode() && hooks.selectLocalTeam != nil {
		hooks.selectLocalTeam(liveTeamID)
	}
	if hooks.afterAttach != nil {
		hooks.afterAttach(team, value, active, netCode, flags, gameBallType)
	}
	hooks.commitMemberCount(team)

	unit := hooks.firstPlayerUnit()
	for unit != zeroObject && hooks.loadNetCode(unit) != netCode {
		unit = hooks.nextPlayerUnit(unit)
	}
	if unit == zeroObject {
		return
	}
	player := hooks.loadPlayer(unit)
	if player == zeroPlayer {
		return
	}
	playerIndex := hooks.loadPlayerIndex(player)
	hooks.resetPlayer(playerIndex)
	hooks.rebuildMasks(playerIndex)
	hooks.markUpdate(unit)
	for owned := hooks.firstOwned(unit); owned != zeroObject; owned = hooks.nextOwned(owned) {
		if hooks.loadClassLow(owned)&0x06 != 0 {
			hooks.markUpdate(owned)
		}
	}
}

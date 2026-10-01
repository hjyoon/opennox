package server

type gameBallPlayerDamageHooks4E1230[O comparable, V, T comparable] struct {
	loadClassLow   func(O) uint8
	loadTypeCache  func() uint32
	lookupType     func(string) uint32
	storeTypeCache func(uint32)
	firstOwned     func(O) O
	nextOwned      func(O) O
	loadType       func(O) uint16
	loadFlags      func(O) uint32
	storeFlags     func(O, uint32)
	applyForce     func(O, O, float32)
	clearOwner     func(O)
	carrierState   func(O, O)
	loadTeam       func(O) V
	hasTeam        func(V) bool
	loadTeamID     func(O) uint8
	findTeam       func(uint8) T
	loadNetCode    func(O) uint32
	changeTeam     func(V, T, uint32, int32)
	createTeam     func(uint8, V, int32, uint32, int32)
	audio          func(uint32, O, int32, uint32)
}

// gameBallPlayerDamage4E1230 preserves GAME.EXE 004E1230, including the
// signed damage threshold, uint32 cache comparison, live reads after the
// release callbacks, and the victim's last-touch record after detaching.
// This is not Wink's 100-force throw, Obj130 clear, or BallStatus 1 operation.
func gameBallPlayerDamage4E1230[O comparable, V, T comparable](source, target O, damage int32, h gameBallPlayerDamageHooks4E1230[O, V, T]) {
	if h.loadClassLow(target)&4 == 0 || damage < 30 {
		return
	}
	typeID := h.loadTypeCache()
	if typeID == 0 {
		typeID = h.lookupType("GameBall")
		h.storeTypeCache(typeID)
	}
	var zero O
	ball := h.firstOwned(target)
	for ball != zero && uint32(h.loadType(ball)) != typeID {
		ball = h.nextOwned(ball)
	}
	if ball == zero {
		return
	}
	h.storeFlags(ball, h.loadFlags(ball)&^0x40)
	h.applyForce(target, ball, 30)
	h.clearOwner(ball)
	h.carrierState(ball, target)
	value := h.loadTeam(ball)
	if h.hasTeam(value) {
		team := h.findTeam(h.loadTeamID(source))
		var noTeam T
		if team != noTeam {
			h.changeTeam(value, team, h.loadNetCode(ball), 0)
		}
	} else {
		// 004E12F1 loads the ball's netcode before the source's team byte.
		netCode := h.loadNetCode(ball)
		h.createTeam(h.loadTeamID(source), value, 1, netCode, 0)
	}
	h.audio(926, target, 0, 0)
}

package server

const (
	ballCollidePlayerClass4EBA00       = uint8(0x04)
	ballCollideNoCollideFlag4EBA00     = uint32(0x40)
	ballCollideDeniedAudio4EBA00       = uint32(928)
	ballCollidePickupAudio4EBA00       = uint32(927)
	ballCollideFeedbackFrames4EBA00    = uint32(45)
	ballCollideTeamStateNormal4EBA00   = uint8(2)
	ballCollideTeamStateSpecial4EBA00  = uint8(4)
	ballCollideSpecialTeamKind4EBA00   = uint8(2)
	ballCollideCantPickupMessage4EBA00 = "objcoll.c:CantPickupBall"
)

type ballCollideHooks4EBA00[O, T comparable, D any] struct {
	loadUpdateData    func(O) D
	loadTeamID        func(O) uint8
	findTeamByID      func(uint8) T
	loadCarrier       func(D) O
	teamMemberCount   func(T) int32
	loadFrame         func() uint32
	loadFeedbackFrame func() uint32
	priorityMessage   func(O, string, int32)
	storeFeedback     func(uint32)
	audio             func(uint32, O)
	loadOwner         func(O) O
	loadClassLow      func(O) uint8
	loadOwnedFirst    func(O) O
	loadTypeInd       func(O) uint16
	loadOwnedNext     func(O) O
	setOwner          func(O, O)
	hasTeam           func(O) int32
	loadNetCode       func(O) uint32
	changeTeam        func(O, T, uint32, int32)
	createTeam        func(uint8, O, int32, uint32, int32)
	loadTeamKind      func(T) uint8
	loadNetCode16     func(O) uint16
	ballStatus        func(uint8, uint16)
	carrierState      func(O, O) O
	loadFlags         func(O) uint32
	storeFlags        func(O, uint32)
	purgeBuffs        func(O)
}

// ballCollide4EBA00 preserves GAME.EXE 004EBA00. The ball update-data pointer
// is entry-cached. A non-nil target causes a team lookup and cached carrier
// read before the ball owner and target class gates. Repeated pickup by the
// current carrier on a multi-member team produces rate-limited feedback and a
// denial sound. A matching type in the target's owned list returns silently.
// Successful pickup preserves the original owner/team/status operations before
// carrier state, pickup audio, the live NoCollide flag update, and buff purge.
// The registered collision argument is not read.
func ballCollide4EBA00[O, T comparable, D any](
	ball, target O,
	_ any,
	hooks ballCollideHooks4EBA00[O, T, D],
) {
	data := hooks.loadUpdateData(ball)
	var zeroObject O
	var zeroTeam T
	team := zeroTeam

	if target != zeroObject {
		team = hooks.findTeamByID(hooks.loadTeamID(target))
		if hooks.loadCarrier(data) == target &&
			team != zeroTeam &&
			hooks.teamMemberCount(team) > 1 {
			frame := hooks.loadFrame()
			if frame-hooks.loadFeedbackFrame() > ballCollideFeedbackFrames4EBA00 {
				hooks.priorityMessage(target, ballCollideCantPickupMessage4EBA00, 0)
				hooks.storeFeedback(hooks.loadFrame())
			}
			hooks.audio(ballCollideDeniedAudio4EBA00, ball)
			return
		}
	}

	if hooks.loadOwner(ball) != zeroObject ||
		target == zeroObject ||
		hooks.loadClassLow(target)&ballCollidePlayerClass4EBA00 == 0 {
		hooks.audio(ballCollideDeniedAudio4EBA00, ball)
		return
	}

	owned := hooks.loadOwnedFirst(target)
	if owned != zeroObject {
		ballType := hooks.loadTypeInd(ball)
		for owned != zeroObject {
			if hooks.loadTypeInd(owned) == ballType {
				return
			}
			owned = hooks.loadOwnedNext(owned)
		}
	}

	hooks.setOwner(target, ball)
	if hooks.hasTeam(ball) != 0 {
		if team != zeroTeam {
			netCode := hooks.loadNetCode(ball)
			hooks.changeTeam(ball, team, netCode, 0)
		}
	} else {
		netCode := hooks.loadNetCode(ball)
		teamID := hooks.loadTeamID(target)
		hooks.createTeam(teamID, ball, 1, netCode, 0)
	}
	if team != zeroTeam {
		state := ballCollideTeamStateNormal4EBA00
		if hooks.loadTeamKind(team) == ballCollideSpecialTeamKind4EBA00 {
			state = ballCollideTeamStateSpecial4EBA00
		}
		hooks.ballStatus(state, hooks.loadNetCode16(target))
	}

	_ = hooks.carrierState(ball, target)
	hooks.audio(ballCollidePickupAudio4EBA00, ball)
	flags := hooks.loadFlags(ball)
	hooks.storeFlags(ball, flags|ballCollideNoCollideFlag4EBA00)
	hooks.purgeBuffs(target)
}

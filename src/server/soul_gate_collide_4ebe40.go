package server

const (
	soulGateQuestFlag4EBE40       = uint32(0x1000)
	soulGatePlayerClassLow4EBE40  = uint8(0x04)
	soulGateReadyState4EBE40      = uint32(1)
	soulGateAudio4EBE40           = uint32(1005)
	soulGatePointFX4EBE40         = uint32(130)
	soulGatePriorityMessage4EBE40 = "objcoll.c:SoulGateCollide"
)

type soulGateCollideHooks4EBE40[O comparable, D, U, P any] struct {
	loadSourceCollideData func(O) D
	gameFlagsCheck        func(uint32) uint32
	loadTargetClassLow    func(O) uint8
	setQuestMode          func(int32)
	firstPlayerUnit       func() O
	nextPlayerUnit        func(O) O
	loadPlayerUpdate      func(O) U
	loadPlayer            func(U) P
	loadQuestState        func(P) uint32
	loadSoulGate          func(U) O
	loadFrame             func() uint32
	setQuestTimer         func(uint32)
	loadLastUsedFrame     func(D) uint32
	loadFPS               func() uint32
	audio                 func(uint32, O, int32, int32)
	pointFX               func(uint32, O) uint32
	priorityMessage       func(O, string, int32)
	storeSoulGate         func(U, O)
	storeLastUsedFrame    func(D, uint32)
}

// soulGateCollide4EBE40 preserves GAME.EXE 004EBE40. The source collision
// record is cached before the Quest-mode test. Eligible collisions disable the
// current Quest mode, scan every live player unit, optionally refresh the
// global Quest timer, throttle feedback with wrapping frame arithmetic, then
// store the gate on the target and the final live frame in the cached record.
// The registered collision point is not read.
func soulGateCollide4EBE40[O comparable, D, U, P, C any](
	source, target O,
	_ C,
	hooks soulGateCollideHooks4EBE40[O, D, U, P],
) {
	data := hooks.loadSourceCollideData(source)
	if hooks.gameFlagsCheck(soulGateQuestFlag4EBE40) == 0 {
		return
	}

	var zeroObject O
	if target == zeroObject {
		return
	}
	if hooks.loadTargetClassLow(target)&soulGatePlayerClassLow4EBE40 == 0 {
		return
	}

	hooks.setQuestMode(0)
	foundReadyGate := false
	for unit := hooks.firstPlayerUnit(); unit != zeroObject; unit = hooks.nextPlayerUnit(unit) {
		update := hooks.loadPlayerUpdate(unit)
		player := hooks.loadPlayer(update)
		if hooks.loadQuestState(player) == soulGateReadyState4EBE40 &&
			hooks.loadSoulGate(update) != zeroObject {
			foundReadyGate = true
		}
	}
	if !foundReadyGate {
		hooks.setQuestTimer(hooks.loadFrame())
	}

	targetUpdate := hooks.loadPlayerUpdate(target)
	notify := hooks.loadSoulGate(targetUpdate) != source
	if !notify {
		frame := hooks.loadFrame()
		last := hooks.loadLastUsedFrame(data)
		fps := hooks.loadFPS()
		notify = frame-last > fps
	}
	if notify {
		hooks.audio(soulGateAudio4EBE40, source, 0, 0)
		_ = hooks.pointFX(soulGatePointFX4EBE40, source)
		hooks.priorityMessage(target, soulGatePriorityMessage4EBE40, 0)
	}

	hooks.storeSoulGate(targetUpdate, source)
	hooks.storeLastUsedFrame(data, hooks.loadFrame())
}

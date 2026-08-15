package server

import "math"

const (
	ankhPlayerClassLow4EBF40          = uint8(0x04)
	ankhPlayerSlotCount4EBF40         = 5
	ankhHistoryCount4EBF40            = 64
	ankhHistoryLifetime4EBF40         = uint32(240)
	ankhFeedbackDelay4EBF40           = uint64(1500)
	ankhFeedbackAudio4EBF40           = uint32(925)
	ankhAwardAudio4EBF40              = uint32(1004)
	ankhAwardPointFX4EBF40            = uint32(130)
	ankhBalanceKey4EBF40              = "MaxExtraLives"
	ankhTradableType4EBF40            = "AnkhTradable"
	ankhAlreadyAwardedMessage4EBF40   = "objcoll.c:ExtraLifeAlreadyAwarded"
	ankhMaximumReachedMessage4EBF40   = "pickup.c:MaxTradableAnkhsReached"
	ankhAwardedExtraLifeMessage4EBF40 = "objcoll.c:AwardExtraLife"
)

type ankhCollideHooks4EBF40[O comparable, D, U, P any] struct {
	loadSourceInitData func(O) D
	loadTargetClassLow func(O) uint8
	loadTargetUpdate   func(O) U
	loadPlayer         func(U) P
	loadQuestAnkh      func(P, int) O
	storeQuestAnkh     func(P, int, O)

	loadFPS                func() uint32
	loadFrame              func() uint32
	loadRecordFrame        func(D, int) uint32
	loadResetName          func() string
	storeRecordName        func(D, int, string)
	loadResetSerialFirst   func() uint8
	storeRecordSerialFirst func(D, int, uint8)
	storeRecordClass       func(D, int, uint8)
	storeRecordFrame       func(D, int, uint32)
	loadRecordClass        func(D, int) uint8
	loadPlayerClass        func(P) uint8
	loadRecordName         func(D, int) string
	loadPlayerName         func(P) string
	loadRecordSerial       func(D, int) string
	loadPlayerSerial       func(P) string
	storeRecordSerial      func(D, int, string)

	ticks              func() uint64
	loadFeedbackTicks  func() uint64
	priorityMessage    func(O, string, int32)
	audio              func(uint32, O, int32, int32)
	storeFeedbackTicks func(uint64)

	loadBalance      func(string) float32
	floatToInt       func(float32) int32
	loadExtraLives   func(U) int32
	newObject        func(string) O
	callPickup       func(O, O, int32, uint32)
	storeSourceFrame func(O, uint32)
	pointFX          func(uint32, O) uint32
	loadNextIndex    func(D) uint8
	storeNextIndex   func(D, uint8)
}

func ankhStoreFirstFree4EBF40[O comparable, D, U, P any](
	source O,
	update U,
	hooks ankhCollideHooks4EBF40[O, D, U, P],
) {
	player := hooks.loadPlayer(update)
	var zero O
	for i := 0; i < ankhPlayerSlotCount4EBF40; i++ {
		if hooks.loadQuestAnkh(player, i) == zero {
			hooks.storeQuestAnkh(player, i, source)
			return
		}
	}
}

func ankhFeedback4EBF40[O comparable, D, U, P any](
	target O,
	message string,
	hooks ankhCollideHooks4EBF40[O, D, U, P],
) {
	now := hooks.ticks()
	last := hooks.loadFeedbackTicks()
	if now-last <= ankhFeedbackDelay4EBF40 {
		return
	}
	hooks.priorityMessage(target, message, 0)
	hooks.audio(ankhFeedbackAudio4EBF40, target, 0, 0)
	hooks.storeFeedbackTicks(hooks.ticks())
}

// ankhCollide4EBF40 preserves GAME.EXE 004EBF40. The source InitData and
// target UpdateData are cached at their original points. The callback scans
// five native object slots and all 64 award-history records, including the
// original unsigned expiration and feedback timers. Player and history-index
// loads remain deliberately live across callbacks. The registered collision
// point is not read.
func ankhCollide4EBF40[O comparable, D, U, P, C any](
	source, target O,
	_ C,
	hooks ankhCollideHooks4EBF40[O, D, U, P],
) {
	data := hooks.loadSourceInitData(source)
	var zero O
	if target == zero {
		return
	}
	if hooks.loadTargetClassLow(target)&ankhPlayerClassLow4EBF40 == 0 {
		return
	}

	update := hooks.loadTargetUpdate(target)
	player := hooks.loadPlayer(update)
	alreadyStored := false
	for i := 0; i < ankhPlayerSlotCount4EBF40; i++ {
		if hooks.loadQuestAnkh(player, i) == source {
			alreadyStored = true
			break
		}
	}

	for i := 0; i < ankhHistoryCount4EBF40; i++ {
		fps := hooks.loadFPS()
		frame := hooks.loadFrame()
		recordFrame := hooks.loadRecordFrame(data, i)
		if frame-recordFrame > ankhHistoryLifetime4EBF40*fps {
			hooks.storeRecordName(data, i, hooks.loadResetName())
			hooks.storeRecordSerialFirst(data, i, hooks.loadResetSerialFirst())
			hooks.storeRecordClass(data, i, 0)
			hooks.storeRecordFrame(data, i, 0)
		}

		player = hooks.loadPlayer(update)
		if hooks.loadRecordClass(data, i) != hooks.loadPlayerClass(player) {
			continue
		}
		if hooks.loadRecordName(data, i) != hooks.loadPlayerName(player) {
			continue
		}
		player = hooks.loadPlayer(update)
		if hooks.loadRecordSerial(data, i) != hooks.loadPlayerSerial(player) {
			continue
		}

		ankhStoreFirstFree4EBF40(source, update, hooks)
		ankhFeedback4EBF40(target, ankhAlreadyAwardedMessage4EBF40, hooks)
		return
	}

	if alreadyStored {
		ankhFeedback4EBF40(target, ankhAlreadyAwardedMessage4EBF40, hooks)
		return
	}

	maximum := hooks.floatToInt(hooks.loadBalance(ankhBalanceKey4EBF40))
	if hooks.loadExtraLives(update) >= maximum {
		ankhFeedback4EBF40(target, ankhMaximumReachedMessage4EBF40, hooks)
		return
	}

	created := hooks.newObject(ankhTradableType4EBF40)
	if created != zero {
		hooks.callPickup(target, created, 1, 0)
	}
	hooks.storeSourceFrame(source, hooks.loadFrame())
	hooks.audio(ankhAwardAudio4EBF40, source, 0, 0)
	_ = hooks.pointFX(ankhAwardPointFX4EBF40, source)
	hooks.priorityMessage(target, ankhAwardedExtraLifeMessage4EBF40, 0)

	ankhStoreFirstFree4EBF40(source, update, hooks)

	player = hooks.loadPlayer(update)
	index := int(hooks.loadNextIndex(data))
	hooks.storeRecordName(data, index, hooks.loadPlayerName(player))

	player = hooks.loadPlayer(update)
	index = int(hooks.loadNextIndex(data))
	hooks.storeRecordClass(data, index, hooks.loadPlayerClass(player))

	index = int(hooks.loadNextIndex(data))
	player = hooks.loadPlayer(update)
	hooks.storeRecordSerial(data, index, hooks.loadPlayerSerial(player))

	index = int(hooks.loadNextIndex(data))
	hooks.storeRecordFrame(data, index, hooks.loadFrame())

	next := hooks.loadNextIndex(data)
	next++
	hooks.storeNextIndex(data, next)
	if next >= ankhHistoryCount4EBF40 {
		hooks.storeNextIndex(data, 0)
	}
}

// ankhRoundFloat32ToInt32_4EBF40 models nox_float2int at 00419A70 under
// GAME.EXE's default x87 round-to-nearest-even mode. Invalid conversions
// produce the signed integer-indefinite value 0x80000000.
func ankhRoundFloat32ToInt32_4EBF40(value float32) int32 {
	if math.IsNaN(float64(value)) || value >= 2147483648 || value < -2147483648 {
		return math.MinInt32
	}
	return int32(math.RoundToEven(float64(value)))
}

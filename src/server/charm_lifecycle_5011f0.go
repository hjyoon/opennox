package server

import (
	"math"

	"github.com/opennox/libs/types"
)

const (
	charmMonsterClass5011F0    = byte(2)
	charmPlayerClass5011F0     = byte(4)
	charmDisabledFlags5011F0   = uint32(0x8020)
	charmImpossibleFlag5013E0  = uint32(0x2000)
	charmMonsterStatus5013E0   = uint32(0x80)
	charmOwnedSubclass5013E0   = uint32(0x80)
	charmControlledClass5013E0 = uint32(0x100)
	charmQuestGameFlag5013E0   = uint32(0x1000)
	charmAcquireTypeFlag5013E0 = uint32(0x08000000)
	charmSearchRange5011F0     = float32(300)
	charmMaximumDistance5013E0 = float64(300)
	charmConfusedEnchant5011F0 = int32(3)
	charmHeldEnchant501690     = int32(5)
	charmPendingEnchant5011F0  = int32(28)
	charmPendingPower5011F0    = int8(5)
	charmAbortSound5011F0      = int32(16)
	charmNonPlayerOrder5013E0  = uint32(4)
)

const (
	charmConfuseDurationKey5011F0 = "ConfuseEnchantDuration"
	charmSmallDurationKey5011F0   = "CharmSmallDuration"
	charmMediumDurationKey5011F0  = "CharmMediumDuration"
	charmLargeDurationKey5011F0   = "CharmLargeDuration"

	charmNoCreatureMessage5011F0     = "Summon.c:CharmNoCreatureCloseEnough"
	charmNotCharmableMessage5011F0   = "Summon.c:CreatureNotCharmable"
	charmNeedGuideMessage5011F0      = "Summon.c:NeedGuideToCharm"
	charmBrokenDistanceMessage5013E0 = "Summon.c:CharmBrokenDistance"
	charmImpossibleMessage5013E0     = "Summon.c:CreatureControlImpossible"
	charmControlFailedMessage5013E0  = "Summon.c:CreatureControlFailed"
	charmAlreadyOwnedMessage5013E0   = "Summon.c:CreatureAlreadyOwned"
)

type charmStartHooks5011F0[Record, Object, Update, Player comparable] struct {
	loadRecordFlag   func(Record) uint32
	balanceFloat     func(string) float64
	balanceFloatInd  func(string, int32) float64
	floatToInt       func(float32) int32
	loadLevel        func(Record) uint32
	loadTarget       func(Record) Object
	storeTarget      func(Record, Object)
	loadCaster       func(Record) Object
	loadSpell        func(Record) uint32
	spellFlags       func(uint32) uint32
	loadPosition     func(Record) *types.Pointf
	searchTarget     func(*types.Pointf, Object, uint32, float32, int32, Object) Object
	loadClassLow     func(Object) byte
	monitored        func(Object, Object) bool
	loadTypeIndex    func(Object) uint16
	charmable        func(uint16) int32
	loadUpdate       func(Object) Update
	loadPlayer       func(Update) Player
	loadGuideLevel   func(Player, int32) uint32
	privateMessage   func(Object, string, byte)
	audio            func(int32, Object)
	guideSize        func(int32) int32
	recordDword      func(Record) uint32
	frame            func() uint32
	storeDeadline    func(Record, uint32)
	buffApply        func(Object, int32, int16, int8)
	attribution      func(Object, Object)
	durationRayStart func(Record)
}

// charmStart5011F0 preserves GAME.EXE 005011F0's callback/load order while
// keeping duration records, objects, updates, and players at native width.
// The balance results are rounded through binary32 before nox_float2int, and
// only the low byte of guide size selects one of the three duration tables.
func charmStart5011F0[Record, Object, Update, Player comparable](
	record Record,
	h charmStartHooks5011F0[Record, Object, Update, Player],
) int32 {
	if h.loadRecordFlag(record) != 0 {
		duration := h.floatToInt(float32(h.balanceFloat(charmConfuseDurationKey5011F0)))
		level := h.loadLevel(record)
		target := h.loadTarget(record)
		h.buffApply(target, charmConfusedEnchant5011F0, int16(duration), int8(level))
		target = h.loadTarget(record)
		caster := h.loadCaster(record)
		h.attribution(caster, target)
		return 1
	}

	cachedCaster := h.loadCaster(record)
	spellID := h.loadSpell(record)
	flags := h.spellFlags(spellID)
	liveCaster := h.loadCaster(record)
	target := h.searchTarget(
		h.loadPosition(record), liveCaster, flags,
		charmSearchRange5011F0, 0, cachedCaster,
	)
	h.storeTarget(record, target)
	var nilObject Object
	if target == nilObject {
		h.privateMessage(h.loadCaster(record), charmNoCreatureMessage5011F0, 0)
		h.audio(charmAbortSound5011F0, h.loadCaster(record))
		return 1
	}

	if h.loadClassLow(target)&charmMonsterClass5011F0 != 0 &&
		!h.monitored(h.loadCaster(record), target) {
		guideTarget := h.loadTarget(record)
		guide := h.charmable(h.loadTypeIndex(guideTarget))
		caster := h.loadCaster(record)
		if h.loadClassLow(caster)&charmPlayerClass5011F0 != 0 {
			update := h.loadUpdate(caster)
			if guide == 0 {
				h.privateMessage(caster, charmNotCharmableMessage5011F0, 0)
				audioCaster := h.loadCaster(record)
				h.storeTarget(record, nilObject)
				h.audio(charmAbortSound5011F0, audioCaster)
				return 1
			}
			player := h.loadPlayer(update)
			if h.loadGuideLevel(player, guide) == 0 {
				h.storeTarget(record, nilObject)
				h.privateMessage(caster, charmNeedGuideMessage5011F0, 0)
				h.audio(charmAbortSound5011F0, h.loadCaster(record))
				return 1
			}
		}

		var duration int32
		switch byte(h.guideSize(guide)) {
		case 1:
			selector := int32(h.loadLevel(record) - 1)
			duration = h.floatToInt(float32(h.balanceFloatInd(charmSmallDurationKey5011F0, selector)))
		case 2:
			selector := int32(h.loadLevel(record) - 1)
			duration = h.floatToInt(float32(h.balanceFloatInd(charmMediumDurationKey5011F0, selector)))
		case 4:
			selector := int32(h.loadLevel(record) - 1)
			duration = h.floatToInt(float32(h.balanceFloatInd(charmLargeDurationKey5011F0, selector)))
		default:
			duration = int32(h.recordDword(record))
		}
		frame := h.frame()
		target = h.loadTarget(record)
		h.storeDeadline(record, frame+uint32(duration))
		buffDuration := int16(uint16(uint32(duration) + 1))
		h.buffApply(target, charmPendingEnchant5011F0, buffDuration, charmPendingPower5011F0)
		h.durationRayStart(record)
		return 0
	}

	audioCaster := h.loadCaster(record)
	h.storeTarget(record, nilObject)
	h.audio(charmAbortSound5011F0, audioCaster)
	return 1
}

type charmFinishHooks5013E0[Record, Object, MonsterUpdate, MonsterDef, PlayerUpdate, Player comparable] struct {
	loadTarget         func(Record) Object
	loadCaster         func(Record) Object
	loadObjectFlags    func(Object) uint32
	distance           func(Object, Object) float64
	loadClassLow       func(Object) byte
	privateMessage     func(Object, string, byte)
	audio              func(int32, Object)
	loadDeadline       func(Record) uint32
	frame              func() uint32
	loadSubclass       func(Object) uint32
	storeSubclass      func(Object, uint32)
	loadTypeIndex      func(Object) uint16
	charmable          func(uint16) int32
	checkLimit         func(Object, int32) bool
	buffOff            func(Object, int32) int32
	findParent         func(Object) Object
	clearOwner         func(Object)
	setOwner           func(Object, Object)
	hasTeam            func(Object) bool
	loadNetCode        func(Object) uint32
	changeTeam         func(Object, uint32)
	loadMonsterUpdate  func(Object) MonsterUpdate
	loadMonsterStatus  func(MonsterUpdate) uint32
	storeMonsterStatus func(MonsterUpdate, uint32)
	gameFlag           func(uint32) bool
	typeHealthMax      func(uint16) uint16
	unitHP             func(Object) uint16
	loadMonsterDef     func(MonsterUpdate) MonsterDef
	monsterDefIsNil    func(MonsterDef) bool
	monsterQuestMax    func(MonsterDef) uint16
	storeHealthMax     func(Object, uint16)
	setHP              func(Object, uint16)
	loadPlayerUpdate   func(Object) PlayerUpdate
	loadPlayer         func(PlayerUpdate) Player
	loadSummonOrder    func(Player) uint32
	loadPlayerIndex    func(Player) byte
	orderUnit          func(Object, Object, uint32)
	reportAcquire      func(byte, Object)
	markMinimap        func(byte, Object, uint32)
	sendSimpleObject   func(byte, Object)
	questSpawnCleanup  func(Object)
	loadSpell          func(Record) uint32
	spellAudio         func(uint32, int32) int32
}

// charmFinish5013E0 preserves the exact completion state machine from
// GAME.EXE 005013E0. The executable has no charm-all-cheat branch here: the
// stale decompiler-only condition is deliberately not modeled.
func charmFinish5013E0[Record, Object, MonsterUpdate, MonsterDef, PlayerUpdate, Player comparable](
	record Record,
	h charmFinishHooks5013E0[Record, Object, MonsterUpdate, MonsterDef, PlayerUpdate, Player],
) int32 {
	target := h.loadTarget(record)
	var nilObject Object
	if target == nilObject || h.loadObjectFlags(target)&charmDisabledFlags5011F0 != 0 {
		h.audio(charmAbortSound5011F0, h.loadCaster(record))
		return 1
	}

	if distance := h.distance(h.loadCaster(record), target); !math.IsNaN(distance) && distance > charmMaximumDistance5013E0 {
		caster := h.loadCaster(record)
		if h.loadClassLow(caster)&charmPlayerClass5011F0 != 0 {
			h.privateMessage(caster, charmBrokenDistanceMessage5013E0, 0)
		}
		h.audio(charmAbortSound5011F0, h.loadCaster(record))
		return 1
	}
	if h.loadDeadline(record)-1 != h.frame() {
		return 0
	}

	target = h.loadTarget(record)
	if h.loadSubclass(target)&charmImpossibleFlag5013E0 != 0 {
		h.privateMessage(h.loadCaster(record), charmImpossibleMessage5013E0, 0)
		h.audio(charmAbortSound5011F0, h.loadCaster(record))
		return 1
	}
	caster := h.loadCaster(record)
	if h.loadClassLow(caster)&charmPlayerClass5011F0 != 0 {
		guide := h.charmable(h.loadTypeIndex(target))
		if !h.checkLimit(h.loadCaster(record), guide) {
			h.privateMessage(h.loadCaster(record), charmControlFailedMessage5013E0, 0)
			h.audio(charmAbortSound5011F0, h.loadCaster(record))
			return 1
		}
	}

	h.buffOff(h.loadTarget(record), charmPendingEnchant5011F0)
	parent := h.findParent(h.loadTarget(record))
	caster = h.loadCaster(record)
	if parent == caster {
		if h.loadClassLow(caster)&charmPlayerClass5011F0 != 0 {
			h.privateMessage(caster, charmAlreadyOwnedMessage5013E0, 0)
		}
		h.audio(charmAbortSound5011F0, h.loadCaster(record))
		return 1
	}

	h.clearOwner(h.loadTarget(record))
	target = h.loadTarget(record)
	h.setOwner(h.loadCaster(record), target)
	if h.hasTeam(h.loadTarget(record)) {
		target = h.loadTarget(record)
		h.changeTeam(target, h.loadNetCode(target))
	}
	target = h.loadTarget(record)
	monsterUpdate := h.loadMonsterUpdate(target)
	h.storeMonsterStatus(monsterUpdate, h.loadMonsterStatus(monsterUpdate)|charmMonsterStatus5013E0)

	if h.gameFlag(charmQuestGameFlag5013E0) {
		typeMax := h.typeHealthMax(h.loadTypeIndex(h.loadTarget(record)))
		currentHP := h.unitHP(h.loadTarget(record))
		definition := h.loadMonsterDef(monsterUpdate)
		maximum := typeMax
		if !h.monsterDefIsNil(definition) {
			maximum = h.monsterQuestMax(definition)
		}
		h.storeHealthMax(h.loadTarget(record), maximum)
		if currentHP > maximum {
			h.setHP(h.loadTarget(record), maximum)
		}
	}

	caster = h.loadCaster(record)
	if h.loadClassLow(caster)&charmPlayerClass5011F0 != 0 {
		playerUpdate := h.loadPlayerUpdate(caster)
		player := h.loadPlayer(playerUpdate)
		order := h.loadSummonOrder(player)
		h.orderUnit(caster, h.loadTarget(record), order)

		target = h.loadTarget(record)
		h.storeSubclass(target, h.loadSubclass(target)|charmOwnedSubclass5013E0)
		target = h.loadTarget(record)
		h.storeSubclass(target, h.loadSubclass(target)|charmControlledClass5013E0)

		player = h.loadPlayer(playerUpdate)
		index := h.loadPlayerIndex(player)
		h.reportAcquire(index, h.loadTarget(record))
		player = h.loadPlayer(playerUpdate)
		index = h.loadPlayerIndex(player)
		h.markMinimap(index, h.loadTarget(record), 1)
		player = h.loadPlayer(playerUpdate)
		index = h.loadPlayerIndex(player)
		h.sendSimpleObject(index, h.loadTarget(record))
		if h.gameFlag(charmQuestGameFlag5013E0) {
			h.questSpawnCleanup(h.loadTarget(record))
		}
	} else {
		h.orderUnit(caster, h.loadTarget(record), charmNonPlayerOrder5013E0)
	}

	caster = h.loadCaster(record)
	soundID := h.spellAudio(h.loadSpell(record), 1)
	h.audio(soundID, caster)
	return 1
}

type charmCancelHooks501690[Record, Object comparable] struct {
	loadTarget      func(Record) Object
	loadObjectFlags func(Object) uint32
	objectDword     func(Object) int32
	buffOff         func(Object, int32) int32
}

// charmCancel501690 retains the initial target dword on nil/disabled exits,
// then removes enchantments five and twenty-eight from live target reloads.
func charmCancel501690[Record, Object comparable](
	record Record,
	h charmCancelHooks501690[Record, Object],
) int32 {
	target := h.loadTarget(record)
	var nilObject Object
	if target == nilObject {
		return h.objectDword(target)
	}
	if h.loadObjectFlags(target)&charmDisabledFlags5011F0 != 0 {
		return h.objectDword(target)
	}
	h.buffOff(target, charmHeldEnchant501690)
	return h.buffOff(h.loadTarget(record), charmPendingEnchant5011F0)
}

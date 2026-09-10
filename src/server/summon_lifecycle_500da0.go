package server

import (
	"math"

	"github.com/opennox/libs/types"
)

const (
	summonSpellBase500DA0       = uint32(74)
	summonPlayerClass500DA0     = byte(4)
	summonDisabledFlags500DA0   = uint32(0x8020)
	summonGuideGameFlags500DA0  = uint32(0x1200)
	summonFinishGameFlags500DA0 = uint32(0x200)
	summonEffectLimit500DA0     = uint16(0xfde8)
	summonPlacementRange500F40  = float64(50)
	summonTraceFlags500F40      = MapTraceFlags(9)
)

const (
	summonNeedGuideMessage500DA0 = "Summon.c:NeedGuideToSummon"
	summonLimitMessage500DA0     = "Summon.c:CreatureControlFailed"
	summonDurationKey500DA0      = "SummonDuration"
)

// summonPayload500DA0 is the semantic form of the packed 14-byte payload
// written to duration-record bytes 72..85 by GAME.EXE 00500DA0.
type summonPayload500DA0 struct {
	typeID    uint16
	position  types.Pointf
	direction byte
	effectID  uint16
	complete  byte
}

type summonStartHooks500DA0[Record, Object comparable, Update, Player any] struct {
	loadSpell        func(Record) uint32
	loadCaster       func(Record) Object
	loadObjectFlags  func(Object) uint32
	loadRecordFlag   func(Record) uint32
	loadClassLowByte func(Object) byte
	loadUpdate       func(Object) Update
	gameFlags        func(uint32) int32
	loadPlayer       func(Update) Player
	loadGuideLevel   func(Player, int32) uint32
	privateMessage   func(Object, string, byte)
	checkLimit       func(Object, int32) int32
	place            func(Record, *types.Pointf) int32
	loadDirectionLow func(Object) byte
	guideName        func(int32) string
	typeID           func(string) uint16
	loadEffectID     func() uint16
	storeEffectID    func(uint16)
	storePayload     func(Record, summonPayload500DA0)
	guideSize        func(int32) int32
	balanceFloat     func(string, int32) float64
	floatToInt       func(float32) int32
	recordDword      func(Record) uint32
	frame            func() uint32
	storeDeadline    func(Record, uint32)
	sendStartEffect  func(uint16, types.Pointf, byte, uint16, uint16)
}

// summonStart500DA0 preserves the observable load/callback order and PE32
// narrowing of GAME.EXE 00500DA0 while Record and Object remain native-width
// handles. In particular, the player's update pointer is cached before the
// game-mode query, caster fields are reloaded around callbacks, and the
// unsupported guide-size branch retains the original record-address dword.
func summonStart500DA0[Record, Object comparable, Update, Player any](
	record Record,
	h summonStartHooks500DA0[Record, Object, Update, Player],
) int32 {
	guide := int32(h.loadSpell(record) - summonSpellBase500DA0)
	caster := h.loadCaster(record)
	var nilObject Object
	if caster == nilObject {
		return 1
	}
	if h.loadObjectFlags(caster)&summonDisabledFlags500DA0 != 0 {
		return 1
	}
	if h.loadRecordFlag(record) != 0 {
		return 1
	}

	if h.loadClassLowByte(caster)&summonPlayerClass500DA0 != 0 {
		update := h.loadUpdate(caster)
		if h.gameFlags(summonGuideGameFlags500DA0) == 1 {
			player := h.loadPlayer(update)
			if h.loadGuideLevel(player, guide) == 0 {
				h.privateMessage(h.loadCaster(record), summonNeedGuideMessage500DA0, 0)
				return 1
			}
		}
		if byte(h.checkLimit(h.loadCaster(record), guide)) == 0 {
			h.privateMessage(h.loadCaster(record), summonLimitMessage500DA0, 0)
			return 1
		}
	}

	var position types.Pointf
	if h.place(record, &position) == 0 {
		return 1
	}
	direction := h.loadDirectionLow(h.loadCaster(record))
	typeID := h.typeID(h.guideName(guide))
	effectID := h.loadEffectID()
	nextEffectID := effectID + 1
	h.storeEffectID(nextEffectID)
	if nextEffectID >= summonEffectLimit500DA0 {
		h.storeEffectID(0)
	}
	h.storePayload(record, summonPayload500DA0{
		typeID:    typeID,
		position:  position,
		direction: direction,
		effectID:  effectID,
		complete:  0,
	})

	var duration int32
	switch byte(h.guideSize(guide)) {
	case 1:
		duration = h.floatToInt(float32(h.balanceFloat(summonDurationKey500DA0, 0)))
	case 2:
		duration = h.floatToInt(float32(h.balanceFloat(summonDurationKey500DA0, 1)))
	case 4:
		duration = h.floatToInt(float32(h.balanceFloat(summonDurationKey500DA0, 2)))
	default:
		duration = int32(h.recordDword(record))
	}
	deadline := h.frame() + uint32(duration)
	h.storeDeadline(record, deadline)
	h.sendStartEffect(effectID, position, direction, typeID, uint16(duration))
	return 0
}

type summonPlacementHooks500F40[Record, Object, Destination comparable, Update, Player any] struct {
	loadCaster       func(Record) Object
	loadClassLowByte func(Object) byte
	loadObjectX      func(Object) float32
	loadObjectY      func(Object) float32
	loadTargetX      func(Record) float32
	loadTargetY      func(Record) float32
	trace            func(types.Pointf, types.Pointf, MapTraceFlags) int32
	tileAllow        func(*types.Pointf) int32
	storeX           func(Destination, float32)
	storeY           func(Destination, float32)
	loadUpdate       func(Object) Update
	loadPlayer       func(Update) Player
	loadPlayerIndex  func(Player) byte
	inform           func(byte, byte, int32)
}

// summonPlacement500F40 models GAME.EXE 00500F40 without interpreting a
// native Go record as a packed PE32 byte array. It deliberately preserves the
// live reloads after trace/tile callbacks and the original exact-one result
// check for mapTileAllowTeleport.
func summonPlacement500F40[Record, Object, Destination comparable, Update, Player any](
	record Record,
	destination Destination,
	h summonPlacementHooks500F40[Record, Object, Destination, Update, Player],
) int32 {
	var nilRecord Record
	if record == nilRecord {
		return 0
	}
	caster := h.loadCaster(record)
	var nilObject Object
	if caster == nilObject {
		return 0
	}
	var nilDestination Destination
	if destination == nilDestination {
		return 0
	}

	class := h.loadClassLowByte(caster)
	fromX := h.loadObjectX(caster)
	if class&summonPlayerClass500DA0 != 0 {
		fromY := h.loadObjectY(caster)
		toX := h.loadTargetX(record)
		toY := h.loadTargetY(record)

		dx := float64(toX) - float64(fromX)
		dy := float32(float64(toY) - float64(fromY))
		distance := math.Sqrt(dx*dx + float64(dy)*float64(dy))
		distance32 := float32(distance)
		if distance > summonPlacementRange500F40 {
			toX = float32(dx*summonPlacementRange500F40/float64(distance32) + float64(fromX))
			toY = float32(float64(dy)*summonPlacementRange500F40/float64(distance32) + float64(fromY))
		}
		from := types.Pointf{X: fromX, Y: fromY}
		to := types.Pointf{X: toX, Y: toY}
		if h.trace(from, to, summonTraceFlags500F40) != 0 && h.tileAllow(&to) != 1 {
			h.storeX(destination, to.X)
			h.storeY(destination, to.Y)
			return 1
		}

		liveCaster := h.loadCaster(record)
		if h.loadClassLowByte(liveCaster)&summonPlayerClass500DA0 != 0 {
			update := h.loadUpdate(liveCaster)
			player := h.loadPlayer(update)
			h.inform(h.loadPlayerIndex(player), 0, 2)
		}
		return 0
	}

	fromY := h.loadObjectY(caster)
	toX := h.loadTargetX(record)
	toY := h.loadTargetY(record)
	if h.trace(
		types.Pointf{X: fromX, Y: fromY},
		types.Pointf{X: toX, Y: toY},
		summonTraceFlags500F40,
	) != 0 {
		h.storeX(destination, h.loadTargetX(record))
		h.storeY(destination, h.loadTargetY(record))
		return 1
	}
	liveCaster := h.loadCaster(record)
	h.storeX(destination, h.loadObjectX(liveCaster))
	h.storeY(destination, h.loadObjectY(liveCaster))
	return 1
}

type summonFinishHooks5010D0[Record, Object comparable, Update, Player any] struct {
	loadCaster           func(Record) Object
	loadObjectFlags      func(Object) uint32
	loadDeadline         func(Record) uint32
	frame                func() uint32
	loadSpell            func(Record) uint32
	loadClassLowByte     func(Object) byte
	loadUpdate           func(Object) Update
	gameFlags            func(uint32) int32
	loadPlayer           func(Update) Player
	loadGuideLevel       func(Player, int32) uint32
	privateMessage       func(Object, string, byte)
	checkLimit           func(Object, int32) int32
	loadPayloadDirection func(Record) byte
	loadPayloadTypeID    func(Record) uint16
	loadPayloadPosition  func(Record) types.Pointf
	summon               func(uint16, types.Pointf, Object, byte) Object
	audioObject          func(Object)
	storeComplete        func(Record, byte)
}

// summonFinish5010D0 preserves GAME.EXE 005010D0's deadline, player checks,
// caster reloads, summon argument loads, and canonical 0/1 result values.
func summonFinish5010D0[Record, Object comparable, Update, Player any](
	record Record,
	h summonFinishHooks5010D0[Record, Object, Update, Player],
) int32 {
	caster := h.loadCaster(record)
	var nilObject Object
	if caster == nilObject || h.loadObjectFlags(caster)&summonDisabledFlags500DA0 != 0 {
		return 1
	}
	if h.loadDeadline(record)-1 != h.frame() {
		return 0
	}

	guide := int32(h.loadSpell(record) - summonSpellBase500DA0)
	if h.loadClassLowByte(caster)&summonPlayerClass500DA0 != 0 {
		update := h.loadUpdate(caster)
		if h.gameFlags(summonFinishGameFlags500DA0) == 1 {
			player := h.loadPlayer(update)
			if h.loadGuideLevel(player, guide) == 0 {
				h.privateMessage(h.loadCaster(record), summonNeedGuideMessage500DA0, 0)
				return 1
			}
		}
		if byte(h.checkLimit(h.loadCaster(record), guide)) == 0 {
			h.privateMessage(h.loadCaster(record), summonLimitMessage500DA0, 0)
			return 1
		}
	}

	direction := h.loadPayloadDirection(record)
	caster = h.loadCaster(record)
	typeID := h.loadPayloadTypeID(record)
	position := h.loadPayloadPosition(record)
	created := h.summon(typeID, position, caster, direction)
	if created != nilObject {
		h.audioObject(created)
	}
	h.storeComplete(record, 1)
	return 1
}

type summonCancelHooks5011C0[Record comparable] struct {
	loadComplete        func(Record) byte
	loadEffectID        func(Record) uint16
	sendCancelEffect    func(uint16)
	loadPayloadPosition func(Record) types.Pointf
	audioPosition       func(types.Pointf)
}

// summonCancel5011C0 emits the cancellation effect and sound only while the
// packed completion byte remains zero, matching GAME.EXE 005011C0.
func summonCancel5011C0[Record comparable](record Record, h summonCancelHooks5011C0[Record]) {
	if h.loadComplete(record) != 0 {
		return
	}
	effectID := h.loadEffectID(record)
	h.sendCancelEffect(effectID)
	position := h.loadPayloadPosition(record)
	h.audioPosition(position)
}

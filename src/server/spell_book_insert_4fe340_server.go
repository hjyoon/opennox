package server

import (
	"unsafe"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
)

// SpellBookInsertAllocator4FE340 snapshots the allocator selected by the
// outer runtime at the same point where GAME.EXE 004FE340 loads its allocator
// global. The returned entity and every other process-local reference remain
// native-width Go pointers.
type SpellBookInsertAllocator4FE340 func() *MagicEntityClass

// SpellBookInsertRuntime4FE340 supplies the operations and intrusive-list
// globals still owned by the root runtime.
type SpellBookInsertRuntime4FE340 struct {
	CheckSummoned  func(*Object, int32) int32
	CountSlaves    func(*Object, uint32, uint32) int32
	SetPlayerState func(*Object, PlayerState)
	LoadAllocator  func() SpellBookInsertAllocator4FE340
	LoadHead       func() *MagicEntityClass
	StoreHead      func(*MagicEntityClass)
}

type spellBookInsertNativeDeps4FE340 struct {
	manaPreflight   func(*Object, *int32, int32) int32
	checkSummoned   func(*Object, int32) int32
	countSlaves     func(*Object, uint32, uint32) int32
	balanceFloat    func(string) float64
	spellPrecheck   func(*Object, int32) int32
	spellCastGate   func(*Object, int32, int32) int32
	informResult    func(uint8, uint8, int32)
	audioEvent      func(int32, *Object, int32, uint32)
	setPlayerState  func(*Object, int32)
	loadFrame       func() uint32
	loadAllocator   func() SpellBookInsertAllocator4FE340
	loadDefinitions func() *PhonemeLeaf
	loadHead        func() *MagicEntityClass
	storeHead       func(*MagicEntityClass)
}

func spellBookInsertNative4FE340(
	unit *Object,
	sequence *int32,
	count, delay, targetMode int32,
	deps spellBookInsertNativeDeps4FE340,
) int32 {
	return spellBookInsert4FE340(spellBookInsertHooks4FE340[
		*Object,
		*int32,
		*PlayerUpdateData,
		*Player,
		*TradeSession,
		*MagicEntityClass,
		*PhonemeLeaf,
		SpellBookInsertAllocator4FE340,
	]{
		loadUnitArg: func() *Object {
			return unit
		},
		loadUnitFlags: func(unit *Object) uint32 {
			return uint32(unit.ObjFlags)
		},
		loadSequenceArg: func() *int32 {
			return sequence
		},
		loadSpell: func(sequence *int32, index int32) int32 {
			return *(*int32)(unsafe.Add(unsafe.Pointer(sequence), uintptr(index)*4))
		},
		loadUnitClassLow: func(unit *Object) uint8 {
			return uint8(unit.ObjClass)
		},
		loadUpdate: func(unit *Object) *PlayerUpdateData {
			return (*PlayerUpdateData)(unit.UpdateData)
		},
		loadTrade: func(update *PlayerUpdateData) *TradeSession {
			return update.Trade70
		},
		loadPlayer: func(update *PlayerUpdateData) *Player {
			return update.Player
		},
		loadKnownSpell: func(player *Player, spellID int32) uint32 {
			return player.SpellLvl[spellID]
		},
		loadSpellCastStart: func(update *PlayerUpdateData) uint32 {
			return update.SpellCastStart
		},
		loadCountArg: func() int32 {
			return count
		},
		manaPreflight: deps.manaPreflight,
		loadPlayerClass: func(player *Player) uint8 {
			return uint8(player.PlayerClass())
		},
		checkSummoned: deps.checkSummoned,
		countSlaves:   deps.countSlaves,
		balanceFloat:  deps.balanceFloat,
		loadCurTrapsLow: func(update *PlayerUpdateData) uint8 {
			return uint8(update.CurTraps)
		},
		spellPrecheck: deps.spellPrecheck,
		spellCastGate: deps.spellCastGate,
		loadPlayerIndex: func(player *Player) uint8 {
			return player.PlayerInd
		},
		informResult:   deps.informResult,
		audioEvent:     deps.audioEvent,
		setPlayerState: deps.setPlayerState,
		loadFrame:      deps.loadFrame,
		storeCasting: func(update *PlayerUpdateData, value uint8) {
			update.Field47_0 = value
		},
		storeSpellCastStart: func(update *PlayerUpdateData, value uint32) {
			update.SpellCastStart = value
		},
		loadAllocator: deps.loadAllocator,
		allocate: func(allocator SpellBookInsertAllocator4FE340) *MagicEntityClass {
			return allocator()
		},
		loadTargetModeArg: func() int32 {
			return targetMode
		},
		loadDelayArg: func() int32 {
			return delay
		},
		storeEntityUnit: func(entity *MagicEntityClass, unit *Object) {
			entity.Obj4 = unit
		},
		storeTargetMode: func(entity *MagicEntityClass, value uint32) {
			entity.Field48 = value
		},
		storeField36: func(entity *MagicEntityClass, value uint8) {
			entity.Field36 = value
		},
		storeEntityFrame: func(entity *MagicEntityClass, value uint32) {
			entity.Frame40 = value
		},
		storeDelay: func(entity *MagicEntityClass, value uint32) {
			entity.Field44 = value
		},
		loadDefinitions: deps.loadDefinitions,
		storeDefinitions: func(entity *MagicEntityClass, value *PhonemeLeaf) {
			entity.Field32 = value
		},
		storeSpellIndex: func(entity *MagicEntityClass, value uint8) {
			entity.SpellInd28 = value
		},
		storeGlyphMode: func(entity *MagicEntityClass, value uint8) {
			entity.Field29 = value
		},
		storeEntitySpell: func(entity *MagicEntityClass, index, value int32) {
			entity.Spells8[index] = value
		},
		loadHead: deps.loadHead,
		storeNext: func(entity, next *MagicEntityClass) {
			entity.Next52 = next
		},
		storePrev: func(entity, prev *MagicEntityClass) {
			entity.Prev56 = prev
		},
		storeHead: deps.storeHead,
	})
}

func spellBookInsertServerDeps4FE340(
	s *Server,
	runtime SpellBookInsertRuntime4FE340,
) spellBookInsertNativeDeps4FE340 {
	return spellBookInsertNativeDeps4FE340{
		manaPreflight: s.SpellManaPreflight4FCEF0,
		checkSummoned: runtime.CheckSummoned,
		countSlaves:   runtime.CountSlaves,
		balanceFloat: func(key string) float64 {
			return s.Balance.Float(key)
		},
		spellPrecheck: func(unit *Object, spellID int32) int32 {
			return s.SpellPrecheck4FD0E0(unit, spell.ID(spellID))
		},
		spellCastGate: func(unit *Object, spellID, glyphMode int32) int32 {
			return s.CheckPlayerCantCastSpell4FD150(unit, spell.ID(spellID), int(glyphMode))
		},
		informResult: func(playerIndex, code uint8, result int32) {
			_ = s.NetInformTextMsg(ntype.PlayerInd(playerIndex), byte(code), int(result))
		},
		audioEvent: func(id int32, unit *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), unit, int(kind), code)
		},
		setPlayerState: func(unit *Object, state int32) {
			runtime.SetPlayerState(unit, PlayerState(state))
		},
		loadFrame:       s.Frame,
		loadAllocator:   runtime.LoadAllocator,
		loadDefinitions: s.Spells.PhonemeTree,
		loadHead:        runtime.LoadHead,
		storeHead:       runtime.StoreHead,
	}
}

// SpellBookInsert4FE340 validates a spell-book gesture and inserts its cast
// record into the native-width magic-entity queue.
//
//go:noinline
func (s *Server) SpellBookInsert4FE340(
	unit *Object,
	sequence *int32,
	count, delay, targetMode int32,
	runtime SpellBookInsertRuntime4FE340,
) int32 {
	return spellBookInsertNative4FE340(
		unit,
		sequence,
		count,
		delay,
		targetMode,
		spellBookInsertServerDeps4FE340(s, runtime),
	)
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjClass)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Object{}.ObjFlags)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(Player{}.SpellLvl[0])]
	_ = [1]struct{}{}[int(spellBookInsertSpellCount4FE340)-len(Player{}.SpellLvl)]
	_ = [1]struct{}{}[1-unsafe.Sizeof(Player{}.PlayerInd)]
	_ = [1]struct{}{}[4-unsafe.Sizeof(MagicEntityClass{}.Spells8[0])]
)

package opennox

import (
	"encoding/binary"

	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type spellCastByBookNativeDeps4FCB80 struct {
	loadHead      func() *server.MagicEntityClass
	storeHead     func(*server.MagicEntityClass)
	loadAllocator func() alloc.ClassT[server.MagicEntityClass]
	freeFirst     func(alloc.ClassT[server.MagicEntityClass], *server.MagicEntityClass)
	loadFrame     func() uint32
	reportStart   func(ntype.PlayerInd, uint8, uint8)
	loadSettings  func() *server.Settings
	loadPhonemes  func(spell.ID) []spell.Phoneme
	loadSuppress  func() uint32
	broadcast     func(*server.Object, int8)
	advanceLeaf   func(*server.PhonemeLeaf, spell.Phoneme) *server.PhonemeLeaf
	phonemeRoot   func() *server.PhonemeLeaf
	informResult  func(ntype.PlayerInd, uint8, int32)
	chargeMana    func(*server.Object, spell.ID, int32) int32
	audioEvent    func(sound.ID, *server.Object, int32, uint32)
	playerSpell   func(*server.Object)
	castByUser    func(spell.ID, int, *server.Object, *server.SpellAcceptArg)
}

// spellBookSpellNative4FCB80 maps the PE32 record's contiguous spell dwords
// onto the native-width record. Slot five aliases the packed control dword at
// original offset 28; later slots would cross a widened pointer and therefore
// deliberately fault instead of truncating it.
func spellBookSpellNative4FCB80(node *server.MagicEntityClass, index int) int32 {
	if index >= 0 && index < len(node.Spells8) {
		return node.Spells8[index]
	}
	if index == len(node.Spells8) {
		return int32(uint32(node.SpellInd28) | uint32(node.Field29)<<8 | uint32(node.Field30)<<16)
	}
	panic("spell-book spell index crosses a native pointer")
}

// spellBookTrapSpellNative4FCB80 preserves the original one-past-array alias:
// trap slot five is the TrapSpellsCnt dword. A larger index would walk into
// unrelated update fields and is allowed to fault rather than being capped.
func spellBookTrapSpellNative4FCB80(update *server.PlayerUpdateData, index int) int32 {
	if index >= 0 && index < len(update.TrapSpells) {
		return int32(update.TrapSpells[index])
	}
	if index == len(update.TrapSpells) {
		return int32(update.TrapSpellsCnt)
	}
	panic("spell-book trap index exceeds the preserved count alias")
}

func spellBookStoreTrapSpellNative4FCB80(update *server.PlayerUpdateData, index int, value int32) {
	if index >= 0 && index < len(update.TrapSpells) {
		update.TrapSpells[index] = uint32(value)
		return
	}
	if index == len(update.TrapSpells) {
		update.TrapSpellsCnt = uint32(value)
		return
	}
	panic("spell-book trap index exceeds the preserved count alias")
}

func spellCastByBookNativeHooks4FCB80(deps spellCastByBookNativeDeps4FCB80) spellCastByBookHooks4FCB80[
	*server.MagicEntityClass,
	*server.Object,
	*server.PlayerUpdateData,
	*server.Player,
	*server.PhonemeLeaf,
	*server.Settings,
	[]spell.Phoneme,
	alloc.ClassT[server.MagicEntityClass],
	*server.SpellAcceptArg,
] {
	return spellCastByBookHooks4FCB80[
		*server.MagicEntityClass,
		*server.Object,
		*server.PlayerUpdateData,
		*server.Player,
		*server.PhonemeLeaf,
		*server.Settings,
		[]spell.Phoneme,
		alloc.ClassT[server.MagicEntityClass],
		*server.SpellAcceptArg,
	]{
		loadHead: deps.loadHead,
		loadObject: func(node *server.MagicEntityClass) *server.Object {
			return node.Obj4
		},
		loadObjectFlags: func(object *server.Object) uint32 {
			return uint32(object.ObjFlags)
		},
		loadNext: func(node *server.MagicEntityClass) *server.MagicEntityClass {
			return node.Next52
		},
		loadPrev: func(node *server.MagicEntityClass) *server.MagicEntityClass {
			return node.Prev56
		},
		storePrev: func(node, previous *server.MagicEntityClass) {
			node.Prev56 = previous
		},
		storeHead: deps.storeHead,
		storeNext: func(node, next *server.MagicEntityClass) {
			node.Next52 = next
		},
		loadAllocator: deps.loadAllocator,
		freeFirst:     deps.freeFirst,
		loadFrame:     deps.loadFrame,
		loadDeadline: func(node *server.MagicEntityClass) uint32 {
			return node.Frame40
		},
		loadObjectClass: func(object *server.Object) uint32 {
			return uint32(object.ObjClass)
		},
		loadUpdate: func(object *server.Object) *server.PlayerUpdateData {
			return (*server.PlayerUpdateData)(object.UpdateData)
		},
		loadProgress: func(node *server.MagicEntityClass) uint8 {
			return node.Field36
		},
		loadSpellIndex: func(node *server.MagicEntityClass) uint8 {
			return node.SpellInd28
		},
		loadSpellLow: func(node *server.MagicEntityClass, index int) uint8 {
			return uint8(spellBookSpellNative4FCB80(node, index))
		},
		loadSpell: spellBookSpellNative4FCB80,
		loadPlayer: func(update *server.PlayerUpdateData) *server.Player {
			return update.Player
		},
		loadPlayerIndex: func(player *server.Player) uint8 {
			return player.PlayerInd
		},
		reportStart: func(index, message, spellLow uint8) {
			deps.reportStart(ntype.PlayerInd(index), message, spellLow)
		},
		loadLeaf: func(node *server.MagicEntityClass) *server.PhonemeLeaf {
			return node.Field32
		},
		loadLeafSpell: func(leaf *server.PhonemeLeaf) int32 {
			return leaf.Ind
		},
		loadSettings: deps.loadSettings,
		loadPhonemeSequence: func(id int32) []spell.Phoneme {
			return deps.loadPhonemes(spell.ID(id))
		},
		loadPhoneme: func(sequence []spell.Phoneme, progress uint8) uint8 {
			return uint8(sequence[progress])
		},
		loadGestureSuppress: deps.loadSuppress,
		loadBroadcastGesture: func(settings *server.Settings) uint32 {
			return binary.LittleEndian.Uint32(settings.BroadcastGestures62[:])
		},
		broadcastPhoneme: func(object *server.Object, phoneme uint8) {
			deps.broadcast(object, int8(phoneme))
		},
		advanceLeaf: func(leaf *server.PhonemeLeaf, phoneme uint8) *server.PhonemeLeaf {
			return deps.advanceLeaf(leaf, spell.Phoneme(phoneme))
		},
		storeLeaf: func(node *server.MagicEntityClass, leaf *server.PhonemeLeaf) {
			node.Field32 = leaf
		},
		storePlayerLeaf: func(update *server.PlayerUpdateData, leaf *server.PhonemeLeaf) {
			update.SpellPhonemeLeaf = leaf
		},
		storeProgress: func(node *server.MagicEntityClass, progress uint8) {
			node.Field36 = progress
		},
		loadGlyphMode: func(node *server.MagicEntityClass) uint8 {
			return node.Field29
		},
		phonemeRoot: deps.phonemeRoot,
		loadDelay: func(node *server.MagicEntityClass) uint32 {
			return node.Field44
		},
		storeDeadline: func(node *server.MagicEntityClass, deadline uint32) {
			node.Frame40 = deadline
		},
		storeSpellIndex: func(node *server.MagicEntityClass, index uint8) {
			node.SpellInd28 = index
		},
		loadTrapCount: func(update *server.PlayerUpdateData) uint8 {
			return uint8(update.TrapSpellsCnt)
		},
		loadTrapSpell: spellBookTrapSpellNative4FCB80,
		informResult: func(index, code uint8, result int32) {
			deps.informResult(ntype.PlayerInd(index), code, result)
		},
		chargeMana: func(object *server.Object, id, mode int32) int32 {
			return deps.chargeMana(object, spell.ID(id), mode)
		},
		audioEvent: func(id int32, object *server.Object, kind int32, code uint32) {
			deps.audioEvent(sound.ID(id), object, kind, code)
		},
		storeTrapSpell: spellBookStoreTrapSpellNative4FCB80,
		storeTrapCount: func(update *server.PlayerUpdateData, count uint8) {
			update.TrapSpellsCnt = update.TrapSpellsCnt&^0xff | uint32(count)
		},
		loadTargetMode: func(node *server.MagicEntityClass) uint32 {
			return node.Field48
		},
		loadCursorX: func(player *server.Player) int32 {
			return int32(player.CursorVec.X)
		},
		loadCursorY: func(player *server.Player) int32 {
			return int32(player.CursorVec.Y)
		},
		storeCastX: func(update *server.PlayerUpdateData, x int32) {
			update.Field55 = int(x)
		},
		storeCastY: func(update *server.PlayerUpdateData, y int32) {
			update.Field56 = int(y)
		},
		storePlayerTarget: func(player *server.Player, object *server.Object) {
			player.Obj3640 = object
		},
		loadCursorObject: func(update *server.PlayerUpdateData) *server.Object {
			return update.CursorObj
		},
		playerSpell: deps.playerSpell,
		storeCastStart: func(update *server.PlayerUpdateData, value uint32) {
			update.SpellCastStart = value
		},
		storeCasting: func(update *server.PlayerUpdateData, value uint8) {
			update.Field47_0 = value
		},
		castByUser: func(id int32, object *server.Object, arg *server.SpellAcceptArg) {
			deps.castByUser(spell.ID(id), -1, object, arg)
		},
	}
}

func spellCastByBookNative4FCB80(deps spellCastByBookNativeDeps4FCB80) {
	spellCastByBook4FCB80(spellCastByBookNativeHooks4FCB80(deps))
}

// nox_xxx_spellCastByBook_4FCB80 is the sole active native-width spell-book
// queue processor. The obsolete PE32 C body and CGo entrypoint are absent.
//
//go:noinline
func nox_xxx_spellCastByBook_4FCB80() {
	s := noxServer
	spellCastByBookNative4FCB80(spellCastByBookNativeDeps4FCB80{
		loadHead: func() *server.MagicEntityClass {
			return magicEntityHead
		},
		storeHead: func(value *server.MagicEntityClass) {
			magicEntityHead = value
		},
		loadAllocator: func() alloc.ClassT[server.MagicEntityClass] {
			return magicEntityAlloc
		},
		freeFirst: func(class alloc.ClassT[server.MagicEntityClass], node *server.MagicEntityClass) {
			class.FreeObjectFirst(node)
		},
		loadFrame: s.Frame,
		reportStart: func(index ntype.PlayerInd, message, spellLow uint8) {
			packet := [...]byte{message, spellLow}
			s.NetList.AddToMsgListCli(index, netlist.Kind1, packet[:])
		},
		loadSettings: getServerSettings,
		loadPhonemes: func(id spell.ID) []spell.Phoneme {
			return s.Spells.DefByInd(id).Def.Phonemes
		},
		loadSuppress: func() uint32 {
			return uint32(legacy.Get_dword_5d4594_2650652())
		},
		broadcast: func(object *server.Object, phoneme int8) {
			s.PlayerPhonemeBroadcast4FC960(object, phoneme)
		},
		advanceLeaf: func(leaf *server.PhonemeLeaf, phoneme spell.Phoneme) *server.PhonemeLeaf {
			return leaf.Next(phoneme)
		},
		phonemeRoot: s.Spells.PhonemeTree,
		informResult: func(index ntype.PlayerInd, code uint8, result int32) {
			s.NetInformTextMsg(index, code, int(result))
		},
		chargeMana: func(object *server.Object, id spell.ID, mode int32) int32 {
			return magicEntityChargeMana(object, id, mode)
		},
		audioEvent: func(id sound.ID, object *server.Object, kind int32, code uint32) {
			s.Audio.EventObj(id, object, int(kind), code)
		},
		playerSpell: s.PlayerSpell,
		castByUser: func(id spell.ID, level int, object *server.Object, arg *server.SpellAcceptArg) {
			s.nox_xxx_castSpellByUser4FDD20(id, level, object, arg)
		},
	})
}

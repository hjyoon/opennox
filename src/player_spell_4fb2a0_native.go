package opennox

import (
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/ntype"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

type playerSpellNativeDeps4FB2A0 struct {
	phonemeRoot    *server.PhonemeLeaf
	hasGameFlag    func(uint32) bool
	hasSpellFlags  func(spell.ID, things.SpellFlags) bool
	isEnemy        func(*server.Object, *server.Object) bool
	precheck       func(*server.Object, spell.ID) int32
	checkCantCast  func(*server.Object, spell.ID, int32) int32
	informText     func(ntype.PlayerInd, byte, int)
	audioEvent     func(sound.ID, *server.Object, int, uint32)
	chargeMana     func(*server.Object, spell.ID, int32) int32
	castSpell      func(spell.ID, *server.Object, *server.SpellAcceptArg) bool
	refundMana     func(*server.Object, int32)
	setState       func(*server.Object, server.PlayerState)
	unknownMessage func() string
	lineMessage    func(*server.Object, string)
	reportSpell    func(ntype.PlayerInd, spell.ID, byte)
}

func playerSpellNativeHooks4FB2A0(deps playerSpellNativeDeps4FB2A0) playerSpellHooks4FB2A0[
	*server.Object,
	*server.PlayerUpdateData,
	*server.Player,
	*server.PhonemeLeaf,
	string,
] {
	return playerSpellHooks4FB2A0[
		*server.Object,
		*server.PlayerUpdateData,
		*server.Player,
		*server.PhonemeLeaf,
		string,
	]{
		loadUpdateData: func(unit *server.Object) *server.PlayerUpdateData {
			return (*server.PlayerUpdateData)(unit.UpdateData)
		},
		loadLeaf: func(update *server.PlayerUpdateData) *server.PhonemeLeaf {
			return update.SpellPhonemeLeaf
		},
		isRootLeaf: func(leaf *server.PhonemeLeaf) bool {
			return leaf == deps.phonemeRoot
		},
		loadSpellID: func(leaf *server.PhonemeLeaf) int32 {
			return leaf.Ind
		},
		hasGameFlag: deps.hasGameFlag,
		loadCursorObj: func(update *server.PlayerUpdateData) *server.Object {
			return update.CursorObj
		},
		hasSpellFlags: func(id int32, flags uint32) bool {
			return deps.hasSpellFlags(spell.ID(id), things.SpellFlags(flags))
		},
		isEnemy: deps.isEnemy,
		loadPlayer: func(update *server.PlayerUpdateData) *server.Player {
			return update.Player
		},
		loadSpellLevel: func(player *server.Player, id int32) uint32 {
			return player.SpellLvl[id]
		},
		precheck: func(unit *server.Object, id int32) int32 {
			return deps.precheck(unit, spell.ID(id))
		},
		checkCantCast: func(unit *server.Object, id, bypass int32) int32 {
			return deps.checkCantCast(unit, spell.ID(id), bypass)
		},
		loadPlayerInd: func(player *server.Player) uint8 {
			return player.PlayerInd
		},
		informResult: func(index, code uint8, value int32) {
			deps.informText(ntype.PlayerInd(index), byte(code), int(value))
		},
		informSpell: func(index, code uint8, leaf *server.PhonemeLeaf) {
			deps.informText(ntype.PlayerInd(index), byte(code), int(leaf.Ind))
		},
		audioEvent: func(id int32, unit *server.Object, kind, code int32) {
			deps.audioEvent(sound.ID(id), unit, int(kind), uint32(code))
		},
		chargeMana: func(unit *server.Object, id, amount int32) int32 {
			return deps.chargeMana(unit, spell.ID(id), amount)
		},
		loadCastTarget: func(player *server.Player) *server.Object {
			return player.Obj3640
		},
		loadCursorPos: func(player *server.Player) (int32, int32) {
			return int32(player.CursorVec.X), int32(player.CursorVec.Y)
		},
		castSpell: func(id int32, unit *server.Object, arg playerSpellArg4FB2A0[*server.Object]) bool {
			nativeArg, freeArg := alloc.New(server.SpellAcceptArg{})
			defer freeArg()
			nativeArg.Obj = arg.target
			nativeArg.Pos = types.Pointf{X: arg.posX, Y: arg.posY}
			return deps.castSpell(spell.ID(id), unit, nativeArg)
		},
		refundMana: deps.refundMana,
		loadState: func(update *server.PlayerUpdateData) uint8 {
			return uint8(update.State)
		},
		setState: func(unit *server.Object, state uint8) {
			deps.setState(unit, server.PlayerState(state))
		},
		unknownMessage: deps.unknownMessage,
		lineMessage:    deps.lineMessage,
		reportSpell: func(index uint8, id int32, status uint8) {
			deps.reportSpell(ntype.PlayerInd(index), spell.ID(id), byte(status))
		},
	}
}

func playerSpellNative4FB2A0(unit *server.Object, deps playerSpellNativeDeps4FB2A0) {
	playerSpell4FB2A0(unit, playerSpellNativeHooks4FB2A0(deps))
}

func (s *Server) playerSpellDeps4FB2A0() playerSpellNativeDeps4FB2A0 {
	return playerSpellNativeDeps4FB2A0{
		phonemeRoot: s.Spells.PhonemeTree(),
		hasGameFlag: func(flag uint32) bool {
			return noxflags.HasGame(noxflags.GameFlag(flag))
		},
		hasSpellFlags: func(id spell.ID, flags things.SpellFlags) bool {
			return s.Spells.HasFlags(id, flags)
		},
		isEnemy: s.IsEnemyTo,
		precheck: func(unit *server.Object, id spell.ID) int32 {
			return s.SpellPrecheck4FD0E0(unit, id)
		},
		checkCantCast: func(unit *server.Object, id spell.ID, bypass int32) int32 {
			return s.S().CheckPlayerCantCastSpell4FD150(unit, id, int(bypass))
		},
		informText: func(index ntype.PlayerInd, code byte, value int) {
			s.NetInformTextMsg(index, code, value)
		},
		audioEvent: func(id sound.ID, unit *server.Object, kind int, code uint32) {
			s.Audio.EventObj(id, unit, kind, code)
		},
		chargeMana: func(unit *server.Object, id spell.ID, amount int32) int32 {
			return magicEntityChargeMana(unit, id, amount)
		},
		castSpell: func(id spell.ID, unit *server.Object, arg *server.SpellAcceptArg) bool {
			return s.nox_xxx_castSpellByUser4FDD20(id, -1, unit, arg)
		},
		refundMana: func(unit *server.Object, mana int32) {
			sub_4FD030(unit, int(mana))
		},
		setState: func(unit *server.Object, state server.PlayerState) {
			nox_xxx_playerSetState_4FA020(unit, state)
		},
		unknownMessage: func() string {
			return s.Strings().GetStringInFile("SpellUnknown", "plyrspel.c")
		},
		lineMessage: func(unit *server.Object, message string) {
			legacy.Nox_xxx_netSendLineMessage_4D9EB0(unit, message)
		},
		reportSpell: func(index ntype.PlayerInd, id spell.ID, status byte) {
			s.NetReportSpellStat(int(index), id, status)
		},
	}
}

// PlayerSpell restores GAME.EXE 004FB2A0 while keeping every Object, Player,
// update-data, phoneme, and spell-argument pointer at native width.
//
//go:noinline
func (s *Server) PlayerSpell(unit *server.Object) {
	playerSpellNative4FB2A0(unit, s.playerSpellDeps4FB2A0())
}

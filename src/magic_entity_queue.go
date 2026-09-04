package opennox

import (
	"encoding/binary"
	"math"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/player"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const (
	magicEntityQueueCapacity = spellRuntimeMagicClassCapacity4FC9B0
	trySpellPacketSize       = 22

	spellResultOK                    = int32(0)
	spellResultCreatureControlFailed = int32(4)
	spellResultTooManyGlyphs         = int32(5)
	spellResultDuplicateInGlyph      = int32(6)
	spellResultBadSkill              = int32(9)
	spellResultIllegal               = int32(10)
	spellResultNotEnoughMana         = int32(11)
	spellResultNotEnoughManaGlyph    = int32(12)
	spellResultCancelledByWarCry     = int32(15)
)

var (
	magicEntityHead  *server.MagicEntityClass
	magicEntityAlloc alloc.ClassT[server.MagicEntityClass]
)

func magicEntityUnlink(it *server.MagicEntityClass) *server.MagicEntityClass {
	next := it.Next52
	if next != nil {
		next.Prev56 = it.Prev56
	}
	if it.Prev56 == nil {
		magicEntityHead = next
	} else {
		it.Prev56.Next52 = next
	}
	magicEntityAlloc.FreeObjectFirst(it)
	return next
}

func magicEntityNextSpell(it *server.MagicEntityClass) int32 {
	next := int(it.SpellInd28) + 1
	if next >= len(it.Spells8) {
		return 0
	}
	return it.Spells8[next]
}

func magicEntitySummonManaCost(sp spell.ID, u *server.Object) int {
	if u == nil || !u.Class().Has(object.ClassPlayer) {
		return 0
	}
	return int(memmap.Uint32(0x587000, 217668+4*uintptr(sp)))
}

func magicEntityCheckMana(u *server.Object, spells *[5]int32, count int32) bool {
	var sequence *int32
	if spells != nil {
		sequence = &spells[0]
	}
	return noxServer.SpellManaPreflight4FCEF0(u, sequence, count) != 0
}

func magicEntityChargeMana(u *server.Object, sp spell.ID, costType int) int {
	if u == nil || !u.Class().Has(object.ClassPlayer) || sp == 0 {
		return -1
	}
	if noxflags.HasEngine(noxflags.EngineGodMode) {
		return 0
	}
	ud := u.UpdateDataPlayer()
	var cost int
	if server.SpellIsSummon(sp) {
		cost = magicEntitySummonManaCost(sp, u)
	} else {
		cost = noxServer.Spells.ManaCost(sp, costType)
	}
	if int(ud.ManaCur) >= cost {
		legacy.Nox_xxx_playerManaSub_4EEBF0(u, cost)
		return cost
	}
	ud.Field20_0 = uint16(noxServer.Spells.ManaCost(sp, 1))
	ud.Field20_1 = uint16(noxServer.TickRate())
	return -1
}

func magicEntitySpellPrecheck(u *server.Object, sp spell.ID) int32 {
	s := noxServer
	owner := u.FindOwnerChainPlayer()
	def := s.Spells.DefByInd(sp)
	if def == nil || !def.IsEnabled() {
		return spellResultIllegal
	}
	if u.Class().Has(object.ClassPlayer) {
		var classFlag things.SpellFlags
		switch u.UpdateDataPlayer().Player.PlayerClass() {
		case player.Wizard:
			classFlag = things.SpellClassWizard
		case player.Conjurer:
			classFlag = things.SpellClassConjurer
		default:
			return spellResultBadSkill
		}
		flags := s.Spells.Flags(sp)
		if !flags.Has(things.SpellClassAny) && !flags.Has(classFlag) {
			return spellResultBadSkill
		}
		return spellResultOK
	}
	if !server.Sub_57AEE0(sp, owner) {
		return spellResultIllegal
	}
	return spellResultOK
}

func magicEntityInform(pl *server.Player, result int32) {
	noxServer.NetInformTextMsg(pl.PlayerIndex(), 0, int(result))
}

func nox_xxx_spellByBookInsert_4FE340(
	u *server.Object,
	spells [5]int32,
	count, delay, targetMode int32,
) int {
	s := noxServer
	if u.ObjFlags&object.Flags(0x8022) != 0 || count < 0 || count > int32(len(spells)) {
		return 0
	}
	for _, raw := range spells {
		if raw < 0 || raw >= server.SpellsMax {
			return 0
		}
	}
	if !u.Class().Has(object.ClassPlayer) {
		return 0
	}
	ud := u.UpdateDataPlayer()
	if ud.Trade70 != nil {
		return 0
	}
	for _, raw := range spells {
		if raw != 0 && ud.Player.SpellLvl[int(raw)] == 0 {
			return 0
		}
	}
	if ud.SpellCastStart != 0 {
		return 0
	}

	hasGlyph := false
	for i := 0; i < int(count); i++ {
		if spell.ID(spells[i]) == spell.SPELL_GLYPH {
			hasGlyph = true
		}
	}
	if hasGlyph {
		if !magicEntityCheckMana(u, &spells, count) {
			magicEntityInform(ud.Player, spellResultNotEnoughManaGlyph)
			s.Audio.EventObj(sound.SoundManaEmpty, u, 0, 0)
			return 0
		}
		if ud.Player.PlayerClass() == player.Conjurer {
			if !nox_xxx_checkSummonedCreaturesLimit_500D70(u, 5) {
				magicEntityInform(ud.Player, spellResultCreatureControlFailed)
				s.Audio.EventObj(sound.SoundPermanentFizzle, u, 0, 0)
				return 0
			}
			if countBombers(u) >= int(s.Balance.Float("MaxBomberCount")) {
				magicEntityInform(ud.Player, spellResultTooManyGlyphs)
				s.Audio.EventObj(sound.SoundPermanentFizzle, u, 0, 0)
				return 0
			}
		} else if int(ud.CurTraps) >= int(s.Balance.Float("MaxTrapCount")) {
			magicEntityInform(ud.Player, spellResultTooManyGlyphs)
			s.Audio.EventObj(sound.SoundPermanentFizzle, u, 0, 0)
			return 0
		}
		for i := 0; i < int(count); i++ {
			sp := spell.ID(spells[i])
			if res := magicEntitySpellPrecheck(u, sp); res != spellResultOK {
				magicEntityInform(ud.Player, res)
				s.Audio.EventObj(sound.SoundPermanentFizzle, u, 0, 0)
				return 0
			}
			if res := s.CheckPlayerCantCastSpell4FD150(u, sp, 1); res != spellResultOK {
				magicEntityInform(ud.Player, res)
				s.Audio.EventObj(sound.SoundPermanentFizzle, u, 0, 0)
				return 0
			}
		}
	} else {
		sp := spell.ID(spells[0])
		if res := magicEntitySpellPrecheck(u, sp); res != spellResultOK {
			magicEntityInform(ud.Player, res)
			s.Audio.EventObj(sound.SoundPermanentFizzle, u, 0, 0)
			return 0
		}
		if res := s.CheckPlayerCantCastSpell4FD150(u, sp, 0); res != spellResultOK {
			magicEntityInform(ud.Player, res)
			s.Audio.EventObj(sound.SoundPermanentFizzle, u, 0, 0)
			return 0
		}
	}

	nox_xxx_playerSetState_4FA020(u, server.PlayerState2)
	ud.Field47_0 = 1
	ud.SpellCastStart = s.Frame()
	if magicEntityAlloc.Class == nil {
		return 0
	}
	e := magicEntityAlloc.NewObject()
	if e == nil {
		return 0
	}
	e.Obj4 = u
	e.Field48 = uint32(targetMode)
	e.Frame40 = s.Frame()
	e.Field44 = uint32(delay)
	e.Field32 = s.Spells.PhonemeTree()
	for i := range e.Spells8 {
		if i < int(count) {
			e.Spells8[i] = spells[i]
			if spell.ID(spells[i]) == spell.SPELL_GLYPH {
				e.Field29 = 1
			}
		}
	}
	e.Next52 = magicEntityHead
	if magicEntityHead != nil {
		magicEntityHead.Prev56 = e
	}
	magicEntityHead = e
	return 1
}

// nox_xxx_spell_4FE680 cancels queued gestures in a warcry radius while all
// object and intrusive-list links remain native-width pointers.
func nox_xxx_spell_4FE680(source *server.Object, radius float32) {
	s := noxServer
	for it := magicEntityHead; it != nil; {
		u := it.Obj4
		dx := float64(u.PosVec.X - source.PosVec.X)
		dy := float64(u.PosVec.Y - source.PosVec.Y)
		shouldCancel := (!u.Class().Has(object.ClassPlayer) || !source.TeamPtr().SameAs(u.TeamPtr())) &&
			math.Sqrt(dx*dx+dy*dy)+0.1 < float64(radius) &&
			s.MapTraceVision(source, u)
		if !shouldCancel {
			it = it.Next52
			continue
		}
		if u.Class().Has(object.ClassPlayer) {
			ud := u.UpdateDataPlayer()
			ud.SpellCastStart = 0
			ud.Field47_0 = 0
			magicEntityInform(ud.Player, spellResultCancelledByWarCry)
			s.Audio.EventObj(sound.SoundPermanentFizzle, u, 0, 0)
			nox_xxx_playerSetState_4FA020(u, server.PlayerState13)
		}
		it = magicEntityUnlink(it)
	}
}

func decodeTrySpellPacket(data []byte) ([5]int32, byte, bool) {
	var spells [5]int32
	if len(data) < trySpellPacketSize || netmsg.Op(data[0]) != netmsg.MSG_TRY_SPELL {
		return spells, 0, false
	}
	for i := range spells {
		spells[i] = int32(binary.LittleEndian.Uint32(data[1+4*i:]))
	}
	return spells, data[21], true
}

func (s *Server) onPacketTrySpell51BAD0(data []byte, pl *server.Player, u *server.Object) (int, bool) {
	spells, targetMode, ok := decodeTrySpellPacket(data)
	if !ok {
		return 0, false
	}
	allowed := true
	if pl.Field3680&0x1 != 0 {
		s.NetPriMsgToPlayer(u, "GeneralPrint:NoSpellWarningGeneral", 0)
		allowed = false
	}
	if pl.Field3680&0x2 != 0 {
		s.NetPriMsgToPlayer(u, "GeneralPrint:ConjureNoSpellWarning1", 0)
		allowed = false
	}
	if !noxflags.HasGame(noxflags.GameModeCoop) && u.ObjFlags&object.Flags(0x4000) != 0 {
		allowed = false
	}
	if noxflags.HasGame(noxflags.GameModeChat) || !allowed {
		return trySpellPacketSize, true
	}
	count := int32(0)
	for _, sp := range spells {
		if sp != 0 {
			count++
		}
	}
	ud := u.UpdateDataPlayer()
	first := spell.ID(spells[0])
	shouldInsert := count != 1 ||
		!s.Spells.HasFlags(first, things.SpellOffensive) ||
		ud.CursorObj == nil ||
		s.IsEnemyTo(u, ud.CursorObj) ||
		noxflags.HasGame(noxflags.GameModeQuest)
	if shouldInsert && nox_xxx_spellByBookInsert_4FE340(u, spells, count, 3, int32(targetMode)) == 0 && count == 1 {
		for _, raw := range spells {
			if raw != 0 {
				s.NetReportSpellStat(int(pl.PlayerIndex()), spell.ID(raw), 0)
			}
		}
	}
	return trySpellPacketSize, true
}

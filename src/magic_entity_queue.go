package opennox

import (
	"encoding/binary"
	"math"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
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

func magicEntityChargeMana(u *server.Object, sp spell.ID, costType int32) int32 {
	return noxServer.SpellManaCharge4FCF90(u, int32(sp), costType, func(unit *server.Object, cost int32) {
		legacy.Nox_xxx_playerManaSub_4EEBF0(unit, int(cost))
	})
}

func magicEntityInform(pl *server.Player, result int32) {
	noxServer.NetInformTextMsg(pl.PlayerIndex(), 0, int(result))
}

func nox_xxx_spellByBookInsert_4FE340(
	u *server.Object,
	spells *int32,
	count, delay, targetMode int32,
) int32 {
	return noxServer.SpellBookInsert4FE340(
		u,
		spells,
		count,
		delay,
		targetMode,
		server.SpellBookInsertRuntime4FE340{
			CheckSummoned: func(unit *server.Object, limit int32) int32 {
				if nox_xxx_checkSummonedCreaturesLimit_500D70(unit, int(limit)) {
					return 1
				}
				return 0
			},
			CountSlaves: legacy.Nox_xxx_unitCountSlaves_4E7CF0,
			SetPlayerState: func(unit *server.Object, state server.PlayerState) {
				_ = nox_xxx_playerSetState_4FA020(unit, state)
			},
			LoadAllocator: func() server.SpellBookInsertAllocator4FE340 {
				allocator := magicEntityAlloc
				return allocator.NewObject
			},
			LoadHead: func() *server.MagicEntityClass {
				return magicEntityHead
			},
			StoreHead: func(entity *server.MagicEntityClass) {
				magicEntityHead = entity
			},
		},
	)
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
	if shouldInsert && nox_xxx_spellByBookInsert_4FE340(u, &spells[0], count, 3, int32(targetMode)) == 0 && count == 1 {
		for _, raw := range spells {
			if raw != 0 {
				s.NetReportSpellStat(int(pl.PlayerIndex()), spell.ID(raw), 0)
			}
		}
	}
	return trySpellPacketSize, true
}

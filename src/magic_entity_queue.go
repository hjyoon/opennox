package opennox

import (
	"github.com/opennox/libs/spell"

	"github.com/opennox/opennox/v1/legacy"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

const (
	magicEntityQueueCapacity = spellRuntimeMagicClassCapacity4FC9B0

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

func nox_xxx_spellByBookInsert_4FE340(
	u *server.Object,
	spells *int32,
	count, delay, targetMode int32,
) int32 {
	return noxServer.SpellBookInsert4FE340(u, spells, count, delay, targetMode)
}

// SpellBookInsert4FE340 supplies the root-owned callbacks and queue globals to
// the native-width server restoration of GAME.EXE 004FE340.
func (s *Server) SpellBookInsert4FE340(
	u *server.Object,
	spells *int32,
	count, delay, targetMode int32,
) int32 {
	return s.Server.SpellBookInsert4FE340(
		u,
		spells,
		count,
		delay,
		targetMode,
		server.SpellBookInsertRuntime4FE340{
			CheckSummoned: func(unit *server.Object, limit int32) int32 {
				if nox_xxx_checkSummonedCreaturesLimit_500D70(unit, limit) {
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

func nox_xxx_spell_4FE680(source *server.Object, radius float32) {
	noxServer.SpellGestureCancel4FE680(source, radius)
}

// SpellGestureCancel4FE680 supplies the root-owned callbacks and queue globals
// to the native-width server restoration of GAME.EXE 004FE680.
func (s *Server) SpellGestureCancel4FE680(source *server.Object, radius float32) {
	s.Server.SpellGestureCancel4FE680(
		source,
		radius,
		server.SpellGestureCancelRuntime4FE680{
			SetPlayerState: func(object *server.Object, state server.PlayerState) {
				_ = nox_xxx_playerSetState_4FA020(object, state)
			},
			LoadHead: func() *server.MagicEntityClass {
				return magicEntityHead
			},
			StoreHead: func(entity *server.MagicEntityClass) {
				magicEntityHead = entity
			},
			LoadAllocator: func() server.SpellGestureCancelAllocator4FE680 {
				allocator := magicEntityAlloc
				return allocator.FreeObjectFirst
			},
		},
	)
}

func (s *Server) onPacketTrySpell51BAD0(data []byte, pl *server.Player, unit *server.Object) (int, bool) {
	return s.Server.NetworkTrySpell51BAD0(data, pl, unit, nil, nox_xxx_spellByBookInsert_4FE340)
}

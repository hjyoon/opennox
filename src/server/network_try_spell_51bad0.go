package server

import (
	"encoding/binary"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/spell"
	"github.com/opennox/libs/things"

	noxflags "github.com/opennox/opennox/v1/common/flags"
)

const NetworkTrySpellPacketSize51BAD0 = 22

func DecodeTrySpellPacket51BAD0(data []byte) ([5]int32, byte, bool) {
	var spells [5]int32
	if len(data) < NetworkTrySpellPacketSize51BAD0 || netmsg.Op(data[0]) != netmsg.MSG_TRY_SPELL {
		return spells, 0, false
	}
	for i := range spells {
		spells[i] = int32(binary.LittleEndian.Uint32(data[1+4*i:]))
	}
	return spells, data[21], true
}

func (s *Server) NetworkTrySpell51BAD0(data []byte, pl *Player, u *Object, update *PlayerUpdateData, insert func(*Object, *int32, int32, int32, int32) int32) (int, bool) {
	spells, targetMode, ok := DecodeTrySpellPacket51BAD0(data)
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
		return NetworkTrySpellPacketSize51BAD0, true
	}
	count := int32(0)
	for _, sp := range spells {
		if sp != 0 {
			count++
		}
	}
	if update == nil {
		update = u.UpdateDataPlayer()
	}
	first := spell.ID(spells[0])
	shouldInsert := count != 1 ||
		!s.Spells.HasFlags(first, things.SpellOffensive) ||
		update.CursorObj == nil ||
		s.IsEnemyTo(u, update.CursorObj) ||
		noxflags.HasGame(noxflags.GameModeQuest)
	if shouldInsert && insert(u, &spells[0], count, 3, int32(targetMode)) == 0 && count == 1 {
		for _, raw := range spells {
			if raw != 0 {
				s.NetReportSpellStat(int(pl.PlayerIndex()), spell.ID(raw), 0)
			}
		}
	}
	return NetworkTrySpellPacketSize51BAD0, true
}

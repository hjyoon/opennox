package server

import (
	"encoding/binary"

	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/common/ntype"
)

// PlayerQuestKeysReport4D9A3F is the native-width binding of the key loops.
// The caller retains the single Quest admission before the extra-life loop;
// do not recheck game flags here after a life/key report callback.
func (s *Server) PlayerQuestKeysReport4D9A3F(unit *Object, update *PlayerUpdateData) {
	playerQuestKeysReport4D9A3F(unit, update, playerQuestKeysReportHooks4D9A3F[*Object, *PlayerUpdateData, *Player]{
		loadCachedType: func(kind int) uint32 {
			return *memmap.PtrT[uint32](0x5D4594, uintptr(1556324+4*kind))
		},
		lookupType: func(kind int) uint32 {
			return uint32(s.Types.IndByID([2]string{"SilverKey", "GoldKey"}[kind]))
		},
		storeCachedType: func(kind int, value uint32) {
			*memmap.PtrT[uint32](0x5D4594, uintptr(1556324+4*kind)) = value
		},
		playerByIndex: func(index int32) *Player { return s.Players.ByInd(ntype.PlayerInd(index)) },
		firstItem:     func(unit *Object) *Object { return unit.InvFirstItem },
		nextItem:      func(item *Object) *Object { return item.InvNextItem },
		typeInd:       func(item *Object) uint16 { return uint16(item.TypeInd) },
		marker: func(update *PlayerUpdateData, kind int, index int32) uint8 {
			if kind == 0 {
				return update.QuestPlayerFlagsA[index]
			}
			return update.QuestPlayerFlagsB[index]
		},
		storeMarker: func(update *PlayerUpdateData, kind int, index int32, value uint8) {
			if kind == 0 {
				update.QuestPlayerFlagsA[index] = value
			} else {
				update.QuestPlayerFlagsB[index] = value
			}
		},
		report: s.questKeyReport4D9DF0,
	})
}

// questKeyReport4D9DF0 preserves 004D9DF0/004D9E30's reliable five-byte
// F0/22 and F0/23 senders. NetCode is the original object's low +36 WORD,
// not a synthesized net code, and the signed send result is returned intact.
func (s *Server) questKeyReport4D9DF0(recipient int32, unit *Object, kind int, presence uint8) int32 {
	code := uint16(unit.NetCode)
	packet := [5]byte{0xf0, byte(22 + kind), presence}
	binary.LittleEndian.PutUint16(packet[3:], code)
	return int32(s.NetSendPacketXxx0(int(recipient), packet[:], nil, 1))
}

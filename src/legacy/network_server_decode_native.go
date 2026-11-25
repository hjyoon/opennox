package legacy

/*
#include "defs.h"

int nox_xxx_netOnPacketRecvServ_51BAD0_net_sdecode_switch(int playerIndex, unsigned char* data, int size, nox_playerInfo* player, nox_object_t* unit, void* update);
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/server"
)

//export nox_xxx_netOnPacketRecvServ_51BAD0_net_sdecode_switch
func nox_xxx_netOnPacketRecvServ_51BAD0_net_sdecode_switch(_ C.int, data *C.uchar, size C.int, player *C.nox_playerInfo, unit *nox_object_t, update unsafe.Pointer) C.int {
	if data == nil || size <= 0 {
		return -1
	}
	packet := unsafe.Slice((*byte)(unsafe.Pointer(data)), int(size))
	return C.int(networkServerDecode51BAD0(packet, asPlayerS(player), asObjectS(unit), (*server.PlayerUpdateData)(update)))
}

func networkServerDecode51BAD0(data []byte, player *server.Player, unit *server.Object, update *server.PlayerUpdateData) int {
	if len(data) == 0 {
		return -1
	}
	netDebug := func() bool { return noxflags.HasEngine(noxflags.EngineNetDebug) }
	traceCode := func(code uint16) { nox_xxx_netTestHighBit_578B70(C.uint(code)) }
	switch netmsg.Op(data[0]) {
	case netmsg.MSG_PLAYER_SET_WAYPOINT:
		if len(data) >= 7 {
			x := binary.LittleEndian.Uint16(data[3:5])
			y := binary.LittleEndian.Uint16(data[5:7])
			nox_xxx_playerSetCustomWP_4F79A0(asObjectC(unit), C.float(x), C.float(y))
			return 7
		}
	case netmsg.MSG_TRY_DROP:
		if len(data) >= server.NetworkTryDropPacketSize51BAD0 {
			return int(networkTryDropCall51BAD0((*[server.NetworkTryDropPacketSize51BAD0]byte)(data), unit, update))
		}
	case netmsg.MSG_TRY_GET:
		if len(data) >= server.NetworkTryGetPacketSize51BAD0 {
			return int(networkTryGetCall51BAD0((*[server.NetworkTryGetPacketSize51BAD0]byte)(data), unit, update))
		}
	case netmsg.MSG_TRY_USE:
		if len(data) >= server.NetworkTryUsePacketSize51BAD0 {
			return int(networkTryUseCall51BAD0((*[server.NetworkTryUsePacketSize51BAD0]byte)(data), unit, update))
		}
	case netmsg.MSG_TRY_EQUIP:
		if len(data) >= server.NetworkTryEquipPacketSize51BAD0 {
			return int(networkTryEquipCall51BAD0((*[server.NetworkTryEquipPacketSize51BAD0]byte)(data), unit, update))
		}
	case netmsg.MSG_TRY_DEQUIP:
		if len(data) >= server.NetworkTryDequipPacketSize51BAD0 {
			return int(networkTryDequipCall51BAD0((*[server.NetworkTryDequipPacketSize51BAD0]byte)(data), unit, update))
		}
	case netmsg.MSG_TRY_CREATURE_COMMAND:
		if len(data) >= server.NetworkTryCreatureCommandPacketSize51BAD0 {
			return int(GetServer().S().NetworkTryCreatureCommand51BAD0(unit, update,
				(*[server.NetworkTryCreatureCommandPacketSize51BAD0]byte)(data),
				server.NetworkTryCreatureCommandRuntime51BAD0{
					NetDebug: netDebug, TestHighBit: traceCode, OrderUnit: Nox_xxx_orderUnit_533900,
				}))
		}
	case netmsg.MSG_TRY_SPELL:
		if len(data) >= server.NetworkTrySpellPacketSize51BAD0 && unit != nil && update != nil && update.Player != nil {
			root := GetServer()
			n, ok := root.S().NetworkTrySpell51BAD0(data, update.Player, unit, update, root.SpellBookInsert4FE340)
			if ok {
				return n
			}
		}
	case netmsg.MSG_TRY_COLLIDE:
		if len(data) >= server.NetworkTryCollidePacketSize51BAD0 {
			return int(networkTryCollideCall51BAD0((*[server.NetworkTryCollidePacketSize51BAD0]byte)(data), unit, update))
		}
	case netmsg.MSG_NEW_ALIAS:
		if len(data) >= server.NetworkNewAliasPacketSize51BAD0 && player != nil {
			return int(GetServer().S().NetworkNewAlias51BAD0(player, (*[server.NetworkNewAliasPacketSize51BAD0]byte)(data)))
		}
	case netmsg.MSG_REPORT_SECONDARY_WEAPON:
		if len(data) >= server.NetworkReportSecondaryWeaponPacketSize51BAD0 {
			return int(GetServer().S().NetworkReportSecondaryWeapon51BAD0(unit,
				(*[server.NetworkReportSecondaryWeaponPacketSize51BAD0]byte)(data),
				server.NetworkReportSecondaryWeaponRuntime51BAD0{
					NetDebug: netDebug, TestHighBit: traceCode,
					Report: func(owner, item *server.Object) { sub_53AB90(asObjectC(owner), asObjectC(item)) },
				}))
		}
	case netmsg.MSG_INFO_BOOK_DATA:
		if len(data) >= server.NetworkInfoBookPacketSize51BAD0 {
			return int(networkInfoBookCall51BAD0((*[server.NetworkInfoBookPacketSize51BAD0]byte)(data), unit, update))
		}
	case netmsg.MSG_VOTE:
		kind, cancel, name, n := server.DecodeVotePacket51BAD0(data, noxflags.HasGame(noxflags.GameModeQuest))
		if n > 0 {
			if cancel {
				VoteCancel506C90(kind, unit, name)
			} else {
				VoteStart506870(kind, unit, name)
			}
			return n
		}
	case netmsg.MSG_TRADE:
		if len(data) >= 2 {
			return int(game4TradeRuntime().packet(unit, data))
		}
	case netmsg.MSG_GAUNTLET:
		if len(data) >= server.NetworkGauntletPacketSize51BAD0 {
			return int(networkGauntletCall51BAD0((*[server.NetworkGauntletPacketSize51BAD0]byte)(data), unit, update))
		}
	case netmsg.MSG_INVENTORY_FAIL:
		if len(data) >= server.NetworkInventoryFailPacketSize51BAD0 {
			return int(networkInventoryFailCall51BAD0((*[server.NetworkInventoryFailPacketSize51BAD0]byte)(data), unit))
		}
	}
	return -1
}

package legacy

/*
#include "GAME1_1.h"
#include "GAME3_2.h"
#include "GAME3_3.h"
#include "flag_collide_4ea400.h"
*/
import "C"

import (
	"encoding/binary"
	"unsafe"

	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/netlist"
	"github.com/opennox/opennox/v1/server"
)

func flagPickupBroadcastInform4DA180(s *server.Server, packet []byte) {
	for player := s.Players.First(); player != nil; player = s.Players.Next(player) {
		if player.PlayerUnit != nil {
			s.NetList.AddToMsgListCli(player.PlayerIndex(), netlist.Kind1, packet)
		}
	}
}

func flagPickupInformOne4DA180(s *server.Server, code, value uint32) {
	var packet [6]byte
	packet[0] = byte(netmsg.MSG_INFORM)
	packet[1] = byte(code)
	binary.LittleEndian.PutUint32(packet[2:], value)
	flagPickupBroadcastInform4DA180(s, packet[:])
}

func flagPickupInformTwo4DA180(s *server.Server, code, first, second uint32) {
	var packet [10]byte
	packet[0] = byte(netmsg.MSG_INFORM)
	packet[1] = byte(code)
	binary.LittleEndian.PutUint32(packet[2:], first)
	binary.LittleEndian.PutUint32(packet[6:], second)
	flagPickupBroadcastInform4DA180(s, packet[:])
}

func flagPickupFlagBallWinner4D8C40(s *server.Server, team *server.Team) {
	var packet [8]byte
	packet[0] = byte(netmsg.MSG_REPORT_FLAG_BALL_WINNER)
	binary.LittleEndian.PutUint16(packet[1:], uint16(uint8(team.ID())))
	packet[3] = 0
	binary.LittleEndian.PutUint32(packet[4:], s.Frame())
	s.NetSendPacketXxx1(255, packet[:], nil, 1)
}

func flagPickupFlagWinner4D8C80(s *server.Server, team *server.Team, kind uint32) {
	var packet [8]byte
	packet[0] = byte(netmsg.MSG_REPORT_FLAG_WINNER)
	teamID := ^uint16(0)
	if team != nil {
		teamID = uint16(uint8(team.ID()))
	}
	binary.LittleEndian.PutUint16(packet[1:], teamID)
	packet[3] = byte(kind)
	binary.LittleEndian.PutUint32(packet[4:], s.Frame())
	s.NetSendPacketXxx1(255, packet[:], nil, 1)
}

func flagPickupObserverUpdate425CA0(first, second *server.Player) {
	// sub_425CA0 still owns a broad ranking/statistics subsystem whose C body
	// takes ABI32 integer addresses. Keeping this conversion in one callback
	// makes the remaining dependency visible to the cross-architecture audit.
	var firstAddr, secondAddr uintptr
	if first != nil {
		firstAddr = uintptr(first.C())
	}
	if second != nil {
		secondAddr = uintptr(second.C())
	}
	C.sub_425CA0(C.int(firstAddr), C.int(secondAddr))
}

func flagPickupTeamEligible418BC0(s *server.Server, team *server.Team) int32 {
	if team == nil {
		return 0
	}
	var count int32
	for player := s.Players.First(); player != nil; player = s.Players.Next(player) {
		unit := player.PlayerUnit
		if s.ObjectByNetCode != nil {
			unit = s.ObjectByNetCode(player.NetCode())
		}
		if unit != nil && unit.TeamVal.ID == team.ID() {
			count++
		}
	}
	return count
}

func flagPickupBuffPurgeRuntime4EA7A0(s *server.Server) server.FlagPickupBuffPurgeRuntime4EA7A0 {
	return server.FlagPickupBuffPurgeRuntime4EA7A0{
		BuffOff: func(obj *server.Object, enchant server.EnchantID) int32 {
			mask := uint32(1) << uint32(enchant)
			result := int32(mask)
			if obj == nil || obj.Buffs&mask == 0 {
				return result
			}
			obj.SetBuffFlags(obj.Buffs&^mask, func(player *server.Player, flags uint32) {
				Nox_xxx_playerResetProtectionCRC_56F7D0(player.ProtUnitBuffs, int(flags))
			})
			obj.BuffsDur[enchant] = 0
			obj.BuffsPower[enchant] = 0
			if enchant != server.ENCHANT_DEATH && enchant != server.ENCHANT_CROWN {
				s.Audio.EventObj(s.Spells.DefByInd(enchant.Spell()).GetOffSound(), obj, 0, 0)
			}
			return 0
		},
	}
}

func flagPickupCTFRuntime4EA490(srv Server) server.FlagPickupCTFRuntime4EA490 {
	s := srv.S()
	return server.FlagPickupCTFRuntime4EA490{
		GameData: func(key uint32) uint16 {
			return uint16(Nox_xxx_servGamedataGet_40A020(uint16(key)))
		},
		MoveHome: func(obj *server.Object, update *server.FlagUpdateData4EA490) {
			Nox_xxx_unitMove_4E7010(obj, update.Home)
		},
		InformReturn: func(netCode uint32) {
			flagPickupInformOne4DA180(s, 4, netCode)
		},
		InformFlag: func(code, netCode, flagIndex uint32) {
			flagPickupInformTwo4DA180(s, code, netCode, flagIndex)
		},
		FlagStatus: Sub_4E82C0,
		ObserverMode: func() uint32 {
			return uint32(Get_dword_5d4594_2650652())
		},
		ObserverUpdate: flagPickupObserverUpdate425CA0,
		DetachInventory: func(owner, item *server.Object) {
			// The typed wrapper contains the remaining 004ED0C0 ABI32 body.
			Sub_4ED0C0(owner, item)
		},
		CreateAt: func(obj, owner *server.Object, pos types.Pointf) {
			if owner == nil {
				srv.CreateObjectAt(obj, nil, pos)
				return
			}
			srv.CreateObjectAt(obj, owner, pos)
		},
		SetGameFlags: func(flags uint32) {
			noxflags.SetGame(noxflags.GameFlag(flags))
		},
		FlagWinner: func(team *server.Team, kind uint32) {
			flagPickupFlagWinner4D8C80(s, team, kind)
		},
		TeamEligible: func(team *server.Team) int32 {
			return flagPickupTeamEligible418BC0(s, team)
		},
		ForceDrop: func(owner, item *server.Object) {
			objectForceDropCall4ED930(owner, item)
		},
		FinalizeDelete: srv.ObjectDeleteLast,
		InventoryPut: func(owner, item *server.Object, mode int32) {
			// The public wrapper is pointer-typed; its C implementation remains
			// listed in the ABI inventory until inventory restoration lands.
			Nox_xxx_inventoryPutImpl_4F3070(owner, item, int(mode))
		},
		ReportObject: func(recipient uint32, obj *server.Object) {
			// 004D82F0 still constructs several modifier-dependent packets from
			// the raw object layout and is an explicit ABI32 dependency.
			C.sub_4D82F0(C.int(recipient), (*C.uint32_t)(obj.CObj()))
		},
		BuffPurge: flagPickupBuffPurgeRuntime4EA7A0(s),
	}
}

func flagPickupBallRuntime4EA800(srv Server) server.FlagPickupBallRuntime4EA800 {
	s := srv.S()
	return server.FlagPickupBallRuntime4EA800{
		GameData: func(key uint32) uint16 {
			return uint16(Nox_xxx_servGamedataGet_40A020(uint16(key)))
		},
		ObserverMode: func() uint32 {
			return uint32(Get_dword_5d4594_2650652())
		},
		ObserverUpdate: flagPickupObserverUpdate425CA0,
		InformScore: func(code, netCode, teamID uint32) {
			flagPickupInformTwo4DA180(s, code, netCode, teamID)
		},
		SetGameFlags: func(flags uint32) {
			noxflags.SetGame(noxflags.GameFlag(flags))
		},
		FlagBallWinner: func(team *server.Team) {
			flagPickupFlagBallWinner4D8C40(s, team)
		},
		ChangeObjectTeam: Nox_xxx_netChangeTeamMb_419570,
		Ticks:            PlatformTicks,
		MoveTo:           Nox_xxx_unitMove_4E7010,
		BallStatus:       Sub_4E8290,
	}
}

//export sub_4EA400
func sub_4EA400(source, target *C.nox_object_t, collision *C.float) {
	srv := GetServer()
	s := srv.S()
	ctfRuntime := flagPickupCTFRuntime4EA490(srv)
	ballRuntime := flagPickupBallRuntime4EA800(srv)
	s.FlagCollide4EA400(
		asObjectS((*nox_object_t)(source)),
		asObjectS((*nox_object_t)(target)),
		(*types.Pointf)(unsafe.Pointer(collision)),
		server.FlagCollideRuntime4EA400{
			PickupCTF: func(source, target *server.Object, collision *types.Pointf) {
				s.FlagPickupCTF4EA490(source, target, collision, ctfRuntime)
			},
			PickupGameBall: func(source, target *server.Object, collision *types.Pointf) {
				s.FlagPickupBall4EA800(source, target, collision, ballRuntime)
			},
		},
	)
}

//export sub_4EA7A0
func sub_4EA7A0(obj *C.nox_object_t) C.int32_t {
	s := GetServer().S()
	return C.int32_t(s.FlagPickupBuffPurge4EA7A0(
		asObjectS((*nox_object_t)(obj)),
		flagPickupBuffPurgeRuntime4EA7A0(s),
	))
}

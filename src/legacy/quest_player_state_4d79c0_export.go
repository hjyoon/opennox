package legacy

/*
#include "defs.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/common/memmap"
	"github.com/opennox/opennox/v1/server"
)

const (
	questPlayerStateBase4D79C0            = uintptr(0x5D4594)
	questPlayerStateTimestampOffset4D7A60 = uintptr(1556172)
)

func questPlayerStateRuntime4D79C0() server.QuestPlayerStateRuntime4D79C0 {
	s := GetServer().S()
	return server.QuestPlayerStateRuntime4D79C0{
		LoadTimestamp: func(index int) uint32 {
			return memmap.Uint32(
				questPlayerStateBase4D79C0,
				questPlayerStateTimestampOffset4D7A60+4*uintptr(index),
			)
		},
		StoreTimestamp: func(index int, value uint32) {
			*memmap.PtrUint32(
				questPlayerStateBase4D79C0,
				questPlayerStateTimestampOffset4D7A60+4*uintptr(index),
			) = value
		},
		Notify: s.QuestNotifyPlayer4D9D20,
		Reset:  Sub_4D6000,
	}
}

func questPlayerStateRemove4D79C0(unit *server.Object) int32 {
	s := GetServer().S()
	return s.QuestPlayerStateRemove4D79C0(unit, questPlayerStateRuntime4D79C0())
}

func questPlayerStateMark4D7A60(index int) int32 {
	s := GetServer().S()
	return s.QuestPlayerStateMark4D7A60(index, questPlayerStateRuntime4D79C0())
}

func questPlayerStateExpire4D7A80() int32 {
	s := GetServer().S()
	return s.QuestPlayerStateExpire4D7A80(questPlayerStateRuntime4D79C0())
}

func questPlayerStateClear4D7B40() int32 {
	s := GetServer().S()
	return s.QuestPlayerStateClear4D7B40(questPlayerStateRuntime4D79C0())
}

//export nox_xxx_questPlayerStateRemove_native_4D79C0
func nox_xxx_questPlayerStateRemove_native_4D79C0(unit *nox_object_t) C.int {
	return C.int(questPlayerStateRemove4D79C0(asObjectS(unit)))
}

//export nox_xxx_questPlayerStateMark_native_4D7A60
func nox_xxx_questPlayerStateMark_native_4D7A60(index C.int) C.int {
	return C.int(questPlayerStateMark4D7A60(int(index)))
}

//export nox_xxx_questPlayerStateExpire_native_4D7A80
func nox_xxx_questPlayerStateExpire_native_4D7A80() C.int {
	return C.int(questPlayerStateExpire4D7A80())
}

//export nox_xxx_questPlayerStateClear_native_4D7B40
func nox_xxx_questPlayerStateClear_native_4D7B40() C.int {
	return C.int(questPlayerStateClear4D7B40())
}

//export nox_xxx_netSendInterestingId_native_4D7BE0
func nox_xxx_netSendInterestingId_native_4D7BE0(unit *nox_object_t) C.int {
	GetServer().S().NetSendInterestingIDOff(asObjectS(unit))
	return 0
}

//export nox_xxx_playerInterestingReset_native_4D7E50
func nox_xxx_playerInterestingReset_native_4D7E50(unit *nox_object_t) C.int {
	GetServer().S().Sub_4D7E50(asObjectS(unit))
	return 0
}

//export nox_xxx_playersInterestingReset_native_4D7EA0
func nox_xxx_playersInterestingReset_native_4D7EA0() C.int {
	GetServer().S().Sub_4D7EA0()
	return 0
}

//export nox_xxx_questNotifyPlayer_native_4D9D20
func nox_xxx_questNotifyPlayer_native_4D9D20(recipient C.int, unit *nox_object_t) C.int {
	return C.int(GetServer().S().QuestNotifyPlayer4D9D20(int(recipient), asObjectS(unit)))
}

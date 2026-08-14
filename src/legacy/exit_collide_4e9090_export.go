package legacy

/*
#include "GAME3_2.h"
#include "GAME3_3.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/server"
)

func storeQuestNextMap4E9090(name string) {
	dst := unsafe.Slice((*byte)(unsafe.Pointer(C.sub_4E8E50())), len(name)+1)
	copy(dst, name)
	dst[len(name)] = 0
}

//export nox_xxx_collideExit_4E9090
func nox_xxx_collideExit_4E9090(
	exit, unit *C.nox_object_t,
	collision *C.float,
) {
	outer := GetServer()
	srv := outer.S()
	srv.ExitCollide4E9090(
		asObjectS((*nox_object_t)(exit)),
		asObjectS((*nox_object_t)(unit)),
		unsafe.Pointer(collision),
		server.ExitCollideRuntime4E9090{
			WarpEnabled: func() int32 {
				return int32(C.sub_4D75E0())
			},
			SaveBusy: func() int32 {
				return int32(Sub_4DCC90())
			},
			ExitAllowed: func(obj *server.Object) int32 {
				return int32(Sub_4DCC10(obj))
			},
			SetMapLoadRequired: func(value int32) {
				Sub_4DCBF0(int(value))
			},
			SetSaveFileName: Nox_setSaveFileName_4DB130,
			SaveCoop: func(enabled int32, obj *server.Object, value int32) {
				Sub_4DB170(enabled != 0, unsafe.Pointer(obj), int(value))
			},
			QuestMapFile: func() unsafe.Pointer {
				return unsafe.Pointer(C.nox_xxx_getQuestMapFile_4D0F60())
			},
			DisableAbility: func(obj *server.Object, ability server.Ability) {
				Sub_4FC300(obj, int(ability))
			},
			CurrentQuestStage: func() uint32 {
				return uint32(C.nox_game_getQuestStage_4E3CC0())
			},
			NextStageThreshold: func(stage uint32) uint32 {
				return uint32(Nox_server_questNextStageThreshold_4D74F0(int32(stage)))
			},
			SendQuestStage: func(index uint8, stage uint32) {
				C.sub_4D7450(C.int(index), C.short(stage))
			},
			SetPlayerState: func(obj *server.Object, state server.PlayerState) {
				Nox_xxx_playerSetState_4FA020(obj, state)
			},
			GoObserver: func(player *server.Player, a2, a3 int32) {
				Nox_xxx_playerGoObserver_4E6860(player, int(a2), int(a3))
			},
			StoreWarpFrame: func(frame uint32) {
				C.sub_4D75F0(C.int(frame))
			},
			StoreNextMap: storeQuestNextMap4E9090,
			SetCurrentQuestStage: func(stage uint32) {
				C.nox_game_setQuestStage_4E3CD0(C.int(stage))
			},
			SetQuestWarping: func(value int32) {
				C.sub_4D76E0(C.int(value))
			},
			MapLoad:   outer.SwitchMap,
			Countdown: questExitCountdownRuntime4E8E60(),
			DelayedDelete: func(obj *server.Object) {
				outer.DelayedDelete(obj)
			},
		},
	)
}

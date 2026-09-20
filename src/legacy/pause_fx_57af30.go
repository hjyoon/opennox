package legacy

/*
#include <stdbool.h>

#include "common__gamemech__pausefx.h"
#include "GAME5_2.h"
*/
import "C"

import (
	"github.com/opennox/libs/noxnet/netmsg"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

const (
	pauseFXLevelUpObject57AF30      = "LevelUp"
	pauseFXOblivionObject57AF30     = "OblivionUp"
	pauseFXPointOpcode57AF30        = netmsg.Op(154)
	pauseFXLevelUpSound57AF30       = sound.ID(902)
	pauseFXLevelUpMessage57AF30     = strman.ID("expLevel.c:LevelUP")
	pauseFXLevelUpMessagePath57AF30 = `C:\NoxPost\src\common\GameMech\PauseFX.c`
	pauseFXLevelUpMessageLine57AF30 = 109
	pauseFXDuration57AF30           = uint64(5000)
)

type pauseFXState57AF30 struct {
	active  bool
	timed   bool
	mode    int32
	started uint64
	unit    *server.Object
	effect  *server.Object
}

type pauseFXRuntime57AF30 struct {
	isPaused        func() bool
	newObject       func(string) *server.Object
	createAt        func(*server.Object, *server.Object, types.Pointf)
	freeObject      func(*server.Object)
	pointFX         func(netmsg.Op, types.Pointf)
	audio           func(sound.ID, *server.Object, int, uint32)
	loadString      func(strman.ID, string, int) string
	sendLine        func(*server.Object, string)
	setPlayerState  func(*server.Object, server.PlayerState) bool
	setAnimation    func(*server.Object, uint8)
	setPause        func(bool)
	ticks           func() uint64
	delayedDelete   func(*server.Object)
	isSpellBookOpen func() bool
}

func pauseFXStart57AF30(
	state *pauseFXState57AF30,
	unit *server.Object,
	mode int32,
	runtime pauseFXRuntime57AF30,
) {
	if state.active || runtime.isPaused() {
		return
	}
	if unit != nil {
		state.unit = unit
	}
	unit = state.unit

	var effect *server.Object
	switch mode {
	case 0:
		effect = runtime.newObject(pauseFXLevelUpObject57AF30)
		state.effect = effect
	case 1:
		effect = runtime.newObject(pauseFXOblivionObject57AF30)
		state.effect = effect
	default:
		effect = state.effect
	}
	if effect != nil {
		if unit != nil {
			runtime.createAt(effect, nil, unit.Pos())
		} else {
			runtime.freeObject(effect)
			state.effect = nil
		}
	}

	if (mode == 0 || mode == 1) && unit != nil {
		runtime.pointFX(pauseFXPointOpcode57AF30, unit.Pos())
	}
	if mode == 0 && unit != nil {
		runtime.audio(pauseFXLevelUpSound57AF30, unit, 2, unit.NetCode)
		message := runtime.loadString(
			pauseFXLevelUpMessage57AF30,
			pauseFXLevelUpMessagePath57AF30,
			pauseFXLevelUpMessageLine57AF30,
		)
		runtime.sendLine(unit, message)
	}
	if unit != nil && runtime.setPlayerState(unit, server.PlayerState30) {
		runtime.setAnimation(unit, 4)
	}

	state.timed = mode == 0 || mode == 1
	state.mode = mode
	state.active = true
	runtime.setPause(true)
	state.started = runtime.ticks()
}

func pauseFXFinish57B0A0(state *pauseFXState57AF30, runtime pauseFXRuntime57AF30) {
	if !state.active {
		return
	}
	unit := state.unit
	if unit != nil && (state.mode == 0 || state.mode == 1) {
		runtime.pointFX(pauseFXPointOpcode57AF30, unit.Pos())
	}
	if state.effect != nil {
		runtime.delayedDelete(state.effect)
	}
	state.effect = nil
	if unit != nil {
		runtime.setPlayerState(unit, server.PlayerState13)
	}
	state.unit = nil
	if !runtime.isSpellBookOpen() {
		runtime.setPause(false)
	}
	state.active = false
}

func pauseFXExpired57B140(state *pauseFXState57AF30, now uint64) bool {
	return state.timed && state.started+pauseFXDuration57AF30 < now
}

func pauseFXNativeRuntime57AF30() pauseFXRuntime57AF30 {
	return pauseFXRuntime57AF30{
		isPaused: func() bool {
			return noxflags.HasGame(noxflags.GamePause)
		},
		newObject: func(id string) *server.Object {
			return GetServer().S().NewObjectByTypeID(id)
		},
		createAt: func(effect, owner *server.Object, pos types.Pointf) {
			GetServer().CreateObjectAt(effect, owner, pos)
		},
		freeObject: func(effect *server.Object) {
			GetServer().S().Objs.FreeObject(effect)
		},
		pointFX: func(op netmsg.Op, pos types.Pointf) {
			GetServer().S().Nox_xxx_netSendPointFx_522FF0(op, pos)
		},
		audio: func(id sound.ID, unit *server.Object, kind int, code uint32) {
			GetServer().S().Audio.EventObj(id, unit, kind, code)
		},
		loadString: func(id strman.ID, path string, line int) string {
			_ = line
			return GetServer().S().Strings().GetStringInFile(id, path)
		},
		sendLine: func(unit *server.Object, message string) {
			Nox_xxx_netSendLineMessage_4D9EB0(unit, message)
		},
		setPlayerState: Nox_xxx_playerSetState_4FA020,
		setAnimation: func(unit *server.Object, frame uint8) {
			(*server.PlayerUpdateData)(unit.UpdateData).Field59_0 = frame
		},
		setPause: func(pause bool) {
			if pause {
				Sub_413A00(1)
			} else {
				Sub_413A00(0)
			}
		},
		ticks: PlatformTicks,
		delayedDelete: func(effect *server.Object) {
			GetServer().DelayedDelete(effect)
		},
		isSpellBookOpen: func() bool {
			return Sub_45D9B0() != 0
		},
	}
}

var pauseFXStateNative57AF30 pauseFXState57AF30

var pauseFXStartCall57AF30 = func(unit *server.Object, mode int32) {
	pauseFXStart57AF30(
		&pauseFXStateNative57AF30,
		unit,
		mode,
		pauseFXNativeRuntime57AF30(),
	)
}

var pauseFXFinishCall57B0A0 = func() {
	pauseFXFinish57B0A0(
		&pauseFXStateNative57AF30,
		pauseFXNativeRuntime57AF30(),
	)
}

func pauseFXStartExportCall57AF30(unit *server.Object, mode int32) {
	C.sub_57AF30(asObjectC(unit), C.int(mode))
}

//export sub_57AF30
func sub_57AF30(unit *C.nox_object_t, mode C.int) {
	pauseFXStartCall57AF30(
		asObjectS((*nox_object_t)(unit)),
		int32(mode),
	)
}

//export nox_xxx_get_57AF20
func nox_xxx_get_57AF20() C.int {
	if pauseFXStateNative57AF30.active {
		return 1
	}
	return 0
}

//export sub_57B0A0
func sub_57B0A0() {
	pauseFXFinishCall57B0A0()
}

//export sub_57B140
func sub_57B140() C.bool {
	return C.bool(pauseFXExpired57B140(
		&pauseFXStateNative57AF30,
		PlatformTicks(),
	))
}

//export nox_xxx___Getcvt_57B180
func nox_xxx___Getcvt_57B180() C.longlong {
	return C.longlong(pauseFXStateNative57AF30.started)
}

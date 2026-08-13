package legacy

import (
	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/server"
)

func netReportPlayerStatusNativeWithSend4D8270(obj *server.Object, send func(byte, []byte) int) int {
	return netReportPlayerStatus4D8270(obj, netPlayerStatusHooks4D8270[*server.Object, *server.PlayerUpdateData, *server.Player]{
		flags: func(obj *server.Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		updateData: func(obj *server.Object) *server.PlayerUpdateData {
			return (*server.PlayerUpdateData)(obj.UpdateData)
		},
		player: func(updateData *server.PlayerUpdateData) *server.Player {
			return updateData.Player
		},
		playerInd: func(player *server.Player) byte {
			return player.PlayerInd
		},
		send: send,
	})
}

type unitFreezeNativeDeps4E79C0 struct {
	reportStatus  func(*server.Object) byte
	setPlayerIdle func(*server.Object)
	raiseZero     func(*server.Object)
	resetPaths    func()
	pushIdle      func(*server.Object) byte
	popAction     func(*server.Object) byte
}

func unitFreezeNative4E79C0(obj *server.Object, source uint32, gate *uint32, deps unitFreezeNativeDeps4E79C0) byte {
	return unitFreeze4E79C0(obj, source, unitFreezeNativeHooks4E79C0(gate, deps))
}

func unitUnfreezeNative4E7A60(obj *server.Object, force uint32, gate *uint32, deps unitFreezeNativeDeps4E79C0) byte {
	return unitUnfreeze4E7A60(obj, force, unitFreezeNativeHooks4E79C0(gate, deps))
}

func unitFreezeNativeHooks4E79C0(gate *uint32, deps unitFreezeNativeDeps4E79C0) unitFreezeHooks4E79C0[*server.Object] {
	return unitFreezeHooks4E79C0[*server.Object]{
		flags: func(obj *server.Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		setFlags: func(obj *server.Object, flags uint32) {
			obj.ObjFlags = object.Flags(flags)
		},
		class: func(obj *server.Object) uint32 {
			return uint32(obj.ObjClass)
		},
		gate: func() uint32 {
			return *gate
		},
		setGate: func(value uint32) {
			*gate = value
		},
		reportStatus:  deps.reportStatus,
		setPlayerIdle: deps.setPlayerIdle,
		raiseZero:     deps.raiseZero,
		resetPaths:    deps.resetPaths,
		firstOwned: func(obj *server.Object) *server.Object {
			return obj.FirstOwned516()
		},
		nextOwned: func(obj *server.Object) *server.Object {
			return obj.NextOwned512()
		},
		monsterStatus: func(obj *server.Object) (byte, bool) {
			loadByte := byte(uintptr(obj.UpdateData))
			return loadByte, obj.UpdateDataMonster().StatusFlags.Has(object.MonStatusSummoned)
		},
		pushIdle:  deps.pushIdle,
		popAction: deps.popAction,
	}
}

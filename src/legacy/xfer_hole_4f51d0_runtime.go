package legacy

/*
#include "server__script__script.h"
*/
import "C"

import (
	"unsafe"

	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type holeXferNativeDeps4F51D0 struct {
	transferScript    func(*server.ScriptCallback, unsafe.Pointer)
	transferInventory func(uint16, *server.Object, int32) int32
}

func holeXferRuntimeDeps4F51D0() holeXferNativeDeps4F51D0 {
	return holeXferNativeDeps4F51D0{
		transferScript: func(callback *server.ScriptCallback, context unsafe.Pointer) {
			// GAME.EXE ignores the script-handler transfer result here.
			C.nox_xxx_xferReadScriptHandler_4F5580(
				unsafe.Pointer(callback),
				(*C.char)(context),
			)
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func holeXferScriptContextNative4F51D0(scriptData unsafe.Pointer) unsafe.Pointer {
	if scriptData == nil {
		return nil
	}
	return unsafe.Add(scriptData, 128)
}

func holeXferReadWriteNative4F51D0(
	cf *cryptfile.CryptFile,
	pointer unsafe.Pointer,
	size int,
) {
	_, _ = cf.ReadWrite(unsafe.Slice((*byte)(pointer), size))
}

func holeXferNative4F51D0(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps holeXferNativeDeps4F51D0,
) int32 {
	return holeXfer4F51D0(
		object,
		holeXferDeps4F51D0[*server.Object, *server.HoleCollideData, unsafe.Pointer]{
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			loadScriptData: func(object *server.Object) unsafe.Pointer {
				return object.Field189
			},
			loadCollideData: func(object *server.Object) *server.HoleCollideData {
				// Preserve the entry pointer without allocation/class validation.
				return (*server.HoleCollideData)(object.CollideData)
			},
			rwVersion: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			mapReadWrite: func(object *server.Object, mapVersion int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, mapVersion)
			},
			rwField24: func(data *server.HoleCollideData) {
				holeXferReadWriteNative4F51D0(cf, unsafe.Pointer(&data.Field24), 4)
			},
			storeField24: func(data *server.HoleCollideData, value uint32) {
				data.Field24 = value
			},
			transferScript: func(data *server.HoleCollideData, scriptData unsafe.Pointer, _ uintptr) {
				deps.transferScript(&data.Script, holeXferScriptContextNative4F51D0(scriptData))
			},
			rwDestinationXY: func(data *server.HoleCollideData) {
				holeXferReadWriteNative4F51D0(cf, unsafe.Pointer(&data.DestinationX), 8)
			},
			rwDestinationExtent: func(data *server.HoleCollideData) {
				holeXferReadWriteNative4F51D0(cf, unsafe.Pointer(&data.DestinationExtent), 4)
			},
			rwDestinationNetCode: func(data *server.HoleCollideData) {
				holeXferReadWriteNative4F51D0(cf, unsafe.Pointer(&data.DestinationNetCode), 2)
			},
			storeScriptFunc: func(data *server.HoleCollideData, value int32) {
				data.Script.Func = value
			},
			storeScriptFlags: func(data *server.HoleCollideData, value uint32) {
				data.Script.Flags = value
			},
			storeDestinationExtent: func(data *server.HoleCollideData, value uint32) {
				data.DestinationExtent = value
			},
			storeDestinationNetCode: func(data *server.HoleCollideData, value uint16) {
				data.DestinationNetCode = value
			},
			readOnly: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_XFerHoleNative4F51D0(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return holeXferNative4F51D0(cf, object, holeXferRuntimeDeps4F51D0())
}

package legacy

/*
#include "GAME4.h"
#include "common__crypt.h"
#include "server__script__script.h"
*/
import "C"

import (
	"encoding/binary"
	"io"
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type triggerXferNativeDeps4F4E50 struct {
	transferScript    func(*server.ScriptCallback, unsafe.Pointer)
	initLegacyScript  func(*server.ScriptCallback)
	transferInventory func(uint16, *server.Object, int32) int32
}

func triggerXferRuntimeDeps4F4E50() triggerXferNativeDeps4F4E50 {
	return triggerXferNativeDeps4F4E50{
		transferScript: func(callback *server.ScriptCallback, context unsafe.Pointer) {
			// GAME.EXE ignores the handler transfer result in TriggerXfer.
			C.nox_xxx_xferReadScriptHandler_4F5580(
				unsafe.Pointer(callback),
				(*C.char)(context),
			)
		},
		initLegacyScript: func(callback *server.ScriptCallback) {
			// sub_4F5540 reads the global mode twice and performs no work unless
			// it is exactly one. CryptFile exposes the same stable Boolean mode.
			_ = cryptfile.Global().ReadOnly()
			if !cryptfile.Global().ReadOnly() {
				return
			}
			C.nox_xxx_mapgenMakeScript_502790(
				C.nox_xxx_mapgenGetSomeFile_426A60(),
				(*C.char)(unsafe.Pointer(callback)),
			)
			if !noxflags.HasGame(noxflags.GameFlag23) {
				callback.Func = -1
			}
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func triggerXferCallbackNative4F4E50(
	update *server.TriggerUpdateData,
	callback triggerXferCallback4F4E50,
) *server.ScriptCallback {
	switch callback {
	case triggerXferActivate4F4E50:
		return &update.ScriptActivate
	case triggerXferDeactivate4F4E50:
		return &update.ScriptDeactivate
	case triggerXferCollide4F4E50:
		return &update.ScriptCollide
	default:
		panic("invalid TriggerXfer callback")
	}
}

func triggerXferNative4F4E50(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps triggerXferNativeDeps4F4E50,
) int32 {
	return triggerXfer4F4E50(
		object,
		triggerXferDeps4F4E50[*server.Object, *server.TriggerUpdateData, unsafe.Pointer]{
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			loadUpdateData: func(object *server.Object) *server.TriggerUpdateData {
				// The original loads this pointer once and does not validate the
				// allocation or object class. Keep the native pointer width.
				return (*server.TriggerUpdateData)(object.UpdateData)
			},
			rwVersion: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			mapReadWrite: func(object *server.Object, mapVersion int32) int32 {
				return objectMapReadWriteNative4F4530(cf, object, mapVersion)
			},
			readOnly: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			loadBoxWidth: func(object *server.Object) float32 {
				return object.Shape.Box.W
			},
			loadBoxHeight: func(object *server.Object) float32 {
				return object.Shape.Box.H
			},
			truncFloatDword: triggerXferTruncFloatDword4F4E50,
			rwBoxWidth: func(value int32) int32 {
				return objectReadOldRWI32Native4F4170(cf, value)
			},
			rwBoxHeight: func(value int32) int32 {
				return objectReadOldRWI32Native4F4170(cf, value)
			},
			storeBoxWidth: func(object *server.Object, value float32) {
				object.Shape.Box.W = value
			},
			storeBoxHeight: func(object *server.Object, value float32) {
				object.Shape.Box.H = value
			},
			calcBox: func(object *server.Object) {
				object.Shape.Box.Calc()
			},
			rwLegacyScratch3: func(value uint32) uint32 {
				var buffer [4]byte
				binary.LittleEndian.PutUint32(buffer[:], value)
				_, _ = cf.ReadWrite(buffer[:3])
				return binary.LittleEndian.Uint32(buffer[:])
			},
			rwColor: func(update *server.TriggerUpdateData, index int) {
				update.Colors[index] = objectReadOldRWU8Native4F4170(cf, update.Colors[index])
			},
			rwFlags: func(update *server.TriggerUpdateData) {
				update.Flags = objectReadOldRWU32Native4F4170(cf, update.Flags)
			},
			loadScriptData: func(object *server.Object) unsafe.Pointer {
				return object.Field189
			},
			transferScript: func(update *server.TriggerUpdateData, callback triggerXferCallback4F4E50, scriptData unsafe.Pointer, offset uintptr) {
				deps.transferScript(
					triggerXferCallbackNative4F4E50(update, callback),
					unsafe.Add(scriptData, offset),
				)
			},
			initLegacyScript: func(update *server.TriggerUpdateData, callback triggerXferCallback4F4E50) {
				deps.initLegacyScript(triggerXferCallbackNative4F4E50(update, callback))
			},
			rwLegacyCount: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			seekCurrent: func(offset int32) {
				_ = cf.Seek(int64(offset), io.SeekCurrent)
			},
			rwClassInclude: func(update *server.TriggerUpdateData) {
				update.ClassInclude = objectReadOldRWU32Native4F4170(cf, update.ClassInclude)
			},
			rwClassExclude: func(update *server.TriggerUpdateData) {
				update.ClassExclude = objectReadOldRWU32Native4F4170(cf, update.ClassExclude)
			},
			storeTeamInclude: func(update *server.TriggerUpdateData, value uint8) {
				update.TeamInclude = value
			},
			storeTeamExclude: func(update *server.TriggerUpdateData, value uint8) {
				update.TeamExclude = value
			},
			rwTeamInclude: func(update *server.TriggerUpdateData) {
				update.TeamInclude = objectReadOldRWU8Native4F4170(cf, update.TeamInclude)
			},
			rwTeamExclude: func(update *server.TriggerUpdateData) {
				update.TeamExclude = objectReadOldRWU8Native4F4170(cf, update.TeamExclude)
			},
			rwState: func(update *server.TriggerUpdateData) {
				update.State = objectReadOldRWU8Native4F4170(cf, update.State)
			},
			rwField9: func(update *server.TriggerUpdateData) {
				update.Field9 = objectReadOldRWU8Native4F4170(cf, update.Field9)
			},
			rwField33: func(object *server.Object) {
				object.Field33 = objectReadOldRWU32Native4F4170(cf, object.Field33)
			},
			loadField33: func(object *server.Object) uint32 {
				return object.Field33
			},
			markAnimationFrame: func(object *server.Object, frame uint32) {
				object.MarkAnimFrame(frame)
			},
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_UnitTriggerXferNative4F4E50(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return triggerXferNative4F4E50(cf, object, triggerXferRuntimeDeps4F4E50())
}

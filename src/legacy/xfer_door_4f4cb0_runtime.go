package legacy

/*
#include "GAME1.h"
*/
import "C"

import (
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/server"
)

type doorXferNativeDeps4F4CB0 struct {
	loadDirectionX    func(int32) int32
	loadDirectionY    func(int32) int32
	attachWall        func(*server.Object, int32, int32)
	transferInventory func(uint16, *server.Object, int32) int32
}

func doorXferRuntimeDeps4F4CB0() doorXferNativeDeps4F4CB0 {
	return doorXferNativeDeps4F4CB0{
		loadDirectionX: server.DoorDirectionX,
		loadDirectionY: server.DoorDirectionY,
		attachWall: func(object *server.Object, tileX, tileY int32) {
			C.nox_xxx_doorAttachWall_410360(asObjectC(object), C.int(tileX), C.int(tileY))
		},
		transferInventory: func(version uint16, object *server.Object, count int32) int32 {
			return xferInventoryCall4F3E30(object, version, count)
		},
	}
}

func doorXferNative4F4CB0(
	cf *cryptfile.CryptFile,
	object *server.Object,
	deps doorXferNativeDeps4F4CB0,
) int32 {
	return doorXfer4F4CB0(
		object,
		doorXferDeps4F4CB0[*server.Object, *server.DoorUpdateData]{
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			loadUpdateData: func(object *server.Object) *server.DoorUpdateData {
				// This direct native-width load is intentional. The original transfer
				// does not validate the allocation or the object's class.
				return object.UpdateDataDoor()
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
			loadCurrentDirection: func(update *server.DoorUpdateData) int32 {
				return update.CurrentDirection
			},
			loadLockCode: func(update *server.DoorUpdateData) uint8 {
				return update.LockCode
			},
			loadTargetDirection: func(update *server.DoorUpdateData) int32 {
				return update.TargetDirection
			},
			rwDirection: func(value int32) int32 {
				return objectReadOldRWI32Native4F4170(cf, value)
			},
			rwLockCode: func(value int32) int32 {
				return objectReadOldRWI32Native4F4170(cf, value)
			},
			rwTargetDirection: func(value int32) int32 {
				return objectReadOldRWI32Native4F4170(cf, value)
			},
			storeCurrentDirection: func(update *server.DoorUpdateData, value int32) {
				update.CurrentDirection = value
			},
			storeFractionalDir: func(update *server.DoorUpdateData, value int16) {
				update.FractionalDir = value
			},
			storeTargetDirection: func(update *server.DoorUpdateData, value int32) {
				update.TargetDirection = value
			},
			storeSyncedDirection: func(update *server.DoorUpdateData, value int32) {
				update.SyncedDirection = value
			},
			loadDirectionX: deps.loadDirectionX,
			loadPositionX: func(object *server.Object) float32 {
				return object.PosVec.X
			},
			loadDirectionY: deps.loadDirectionY,
			loadPositionY: func(object *server.Object) float32 {
				return object.PosVec.Y
			},
			truncQwordLow: doorXferTruncSignedQwordLow4F4CB0,
			attachWall:    deps.attachWall,
			storeTileX: func(update *server.DoorUpdateData, value int32) {
				update.TileX = value
			},
			storeTileY: func(update *server.DoorUpdateData, value int32) {
				update.TileY = value
			},
			storeLockCode: func(update *server.DoorUpdateData, value uint8) {
				update.LockCode = value
			},
			transferInventory: deps.transferInventory,
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
		},
	)
}

func Nox_xxx_XFerDoorNative4F4CB0(
	cf *cryptfile.CryptFile,
	object *server.Object,
) int32 {
	return doorXferNative4F4CB0(cf, object, doorXferRuntimeDeps4F4CB0())
}

package legacy

import (
	"unsafe"

	objectlib "github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func objectMapReadWriteNative4F4530(
	cf *cryptfile.CryptFile,
	object *server.Object,
	mapVersion int32,
) int32 {
	return objectMapReadWrite4F4530(object, mapVersion,
		objectMapReadWriteDeps4F4530[*server.Object, unsafe.Pointer]{
			loadField34: func(object *server.Object) uint32 {
				return object.Field34
			},
			readOnly: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			rwU16: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			readOld: func(object *server.Object, objectVersion, mapVersion int32) int32 {
				return objectReadOldNative4F4170(cf, object, objectVersion, mapVersion)
			},
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
			rwExtent: func(object *server.Object) {
				object.Extent = objectReadOldRWU32Native4F4170(cf, object.Extent)
			},
			rwScriptID: func(object *server.Object) {
				object.ScriptIDVal = objectReadOldRWI32Native4F4170(cf, object.ScriptIDVal)
			},
			loadScriptID: func(object *server.Object) int32 {
				return object.ScriptIDVal
			},
			gameFlags: func(mask uint32) int32 {
				if noxflags.HasGame(noxflags.GameFlag(mask)) {
					return 1
				}
				return 0
			},
			nextScriptID: func() int32 {
				return int32(GetServer().S().Objs.NextObjectScriptID())
			},
			storeScriptID: func(object *server.Object, value int32) {
				object.ScriptIDVal = value
			},
			rwPositionX: func(object *server.Object) {
				object.PosVec.X = objectReadOldRWF32Native4F4170(cf, object.PosVec.X)
			},
			rwPositionY: func(object *server.Object) {
				object.PosVec.Y = objectReadOldRWF32Native4F4170(cf, object.PosVec.Y)
			},
			loadPositionX: func(object *server.Object) float32 {
				return object.PosVec.X
			},
			loadPositionY: func(object *server.Object) float32 {
				return object.PosVec.Y
			},
			storeNewPositionX: func(object *server.Object, value float32) {
				object.NewPos.X = value
			},
			storeNewPositionY: func(object *server.Object, value float32) {
				object.NewPos.Y = value
			},
			extendedAdmission: func(object *server.Object) int8 {
				return GetServer().S().Sub_4F40A0(object)
			},
			rwU8: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
			},
			loadFlags: func(object *server.Object) uint32 {
				return uint32(object.ObjFlags)
			},
			rwU32: func(value uint32) uint32 {
				return objectReadOldRWU32Native4F4170(cf, value)
			},
			storeFlags: func(object *server.Object, value uint32) {
				object.ObjFlags = objectlib.Flags(value)
			},
			setOn: func(object *server.Object) {
				objectSetOnRuntime4E75B0(object)
			},
			setOff: func(object *server.Object) {
				objectSetOffRuntime4E7600(object)
			},
			loadIDPointer: func(object *server.Object) unsafe.Pointer {
				return object.IDPtr
			},
			stringLength: func(pointer unsafe.Pointer) uintptr {
				return uintptr(len(alloc.GoString((*byte)(pointer))))
			},
			allocateID: func(size uint16) unsafe.Pointer {
				pointer, _ := alloc.Calloc(1, uintptr(size))
				return pointer
			},
			storeIDPointer: func(object *server.Object, pointer unsafe.Pointer) {
				object.IDPtr = pointer
			},
			rwIDBytes: func(pointer unsafe.Pointer, length uint8) {
				_, _ = cf.ReadWrite(unsafe.Slice((*byte)(pointer), int(length)))
			},
			terminateID: func(pointer unsafe.Pointer, length uint8) {
				*(*byte)(unsafe.Add(pointer, uintptr(length))) = 0
			},
			rwTeamID: func(object *server.Object) {
				object.TeamVal.ID = server.TeamID(objectReadOldRWU8Native4F4170(cf, uint8(object.TeamVal.ID)))
			},
			loadInventoryHead: func(object *server.Object) *server.Object {
				return object.InvFirstItem
			},
			loadInventoryNext: func(object *server.Object) *server.Object {
				return object.InvNextItem
			},
			loadField129: func(object *server.Object) *server.Object {
				return object.Field129
			},
			loadTypeInd: func(object *server.Object) uint16 {
				return object.TypeInd
			},
			ownedTypeAllowed: func(typeInd uint16) int32 {
				if Sub_4E3B80(int(typeInd)) {
					return 1
				}
				return 0
			},
			loadField128: func(object *server.Object) *server.Object {
				return object.Field128
			},
			rwI32: func(value int32) int32 {
				return objectReadOldRWI32Native4F4170(cf, value)
			},
			readOwnedScriptID: func() int32 {
				return objectReadOldRWI32Native4F4170(cf, 0)
			},
			addPendingOwn: func(ownerScriptID, ownedScriptID int32) {
				pendingOwns516EE0.add(ownerScriptID, ownedScriptID)
			},
			rwOwnedScriptID: func(object *server.Object) {
				object.ScriptIDVal = objectReadOldRWI32Native4F4170(cf, object.ScriptIDVal)
			},
			loadField5: func(object *server.Object) uint32 {
				return object.Field5
			},
			unsetStatus: func(object *server.Object, status uint32) {
				object.UnsetXStatus(status)
			},
			setStatus: func(object *server.Object, status uint32) {
				object.SetXStatus(status)
			},
			loadField189: func(object *server.Object) unsafe.Pointer {
				return object.Field189
			},
			scriptHandler: func(object *server.Object, context unsafe.Pointer) int32 {
				return objectMapScriptHandlerNative4F4530(object, context)
			},
			gameFrame: gameFrameHook,
			storeField32: func(object *server.Object, value uint32) {
				object.Field32 = value
			},
		})
}

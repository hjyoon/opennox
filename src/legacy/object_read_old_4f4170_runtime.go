package legacy

import (
	"encoding/binary"
	"math"
	"unsafe"

	objectlib "github.com/opennox/libs/object"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/internal/cryptfile"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
	"github.com/opennox/opennox/v1/server"
)

func objectReadOldRWU8Native4F4170(cf *cryptfile.CryptFile, value uint8) uint8 {
	buf := [1]byte{value}
	_, _ = cf.ReadWrite(buf[:])
	return buf[0]
}

func objectReadOldRWU16Native4F4170(cf *cryptfile.CryptFile, value uint16) uint16 {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], value)
	_, _ = cf.ReadWrite(buf[:])
	return binary.LittleEndian.Uint16(buf[:])
}

func objectReadOldRWU32Native4F4170(cf *cryptfile.CryptFile, value uint32) uint32 {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	_, _ = cf.ReadWrite(buf[:])
	return binary.LittleEndian.Uint32(buf[:])
}

func objectReadOldRWI32Native4F4170(cf *cryptfile.CryptFile, value int32) int32 {
	return int32(objectReadOldRWU32Native4F4170(cf, uint32(value)))
}

func objectReadOldRWF32Native4F4170(cf *cryptfile.CryptFile, value float32) float32 {
	return math.Float32frombits(objectReadOldRWU32Native4F4170(cf, math.Float32bits(value)))
}

func objectReadOldNative4F4170(
	cf *cryptfile.CryptFile,
	object *server.Object,
	objectVersion, mapVersion int32,
) int32 {
	return objectReadOldVer4F4170(object, objectVersion, mapVersion,
		objectReadOldDeps4F4170[*server.Object, unsafe.Pointer]{
			readOnly: func() int32 {
				if cf.ReadOnly() {
					return 1
				}
				return 0
			},
			storeField34: func(object *server.Object, value uint32) {
				object.Field34 = value
			},
			rwExtent: func(object *server.Object) {
				object.Extent = objectReadOldRWU32Native4F4170(cf, object.Extent)
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
			rwPositionX: func(object *server.Object) {
				object.PosVec.X = objectReadOldRWF32Native4F4170(cf, object.PosVec.X)
			},
			rwPositionY: func(object *server.Object) {
				object.PosVec.Y = objectReadOldRWF32Native4F4170(cf, object.PosVec.Y)
			},
			rwOldPosition: func(x, y int32) (int32, int32) {
				var buf [8]byte
				binary.LittleEndian.PutUint32(buf[0:4], uint32(x))
				binary.LittleEndian.PutUint32(buf[4:8], uint32(y))
				_, _ = cf.ReadWrite(buf[:])
				return int32(binary.LittleEndian.Uint32(buf[0:4])), int32(binary.LittleEndian.Uint32(buf[4:8]))
			},
			storePositionX: func(object *server.Object, value float32) {
				object.PosVec.X = value
			},
			storePositionY: func(object *server.Object, value float32) {
				object.PosVec.Y = value
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
			loadIDPointer: func(object *server.Object) unsafe.Pointer {
				return object.IDPtr
			},
			stringLength: func(pointer unsafe.Pointer) uintptr {
				return uintptr(len(alloc.GoString((*byte)(pointer))))
			},
			rwU8: func(value uint8) uint8 {
				return objectReadOldRWU8Native4F4170(cf, value)
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
			rwScriptID: func(object *server.Object) {
				object.ScriptIDVal = objectReadOldRWI32Native4F4170(cf, object.ScriptIDVal)
			},
			loadScriptID: func(object *server.Object) int32 {
				return object.ScriptIDVal
			},
			storeScriptID: func(object *server.Object, value int32) {
				object.ScriptIDVal = value
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
			rwU16: func(value uint16) uint16 {
				return objectReadOldRWU16Native4F4170(cf, value)
			},
			rwI32: func(value int32) int32 {
				return objectReadOldRWI32Native4F4170(cf, value)
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
		})
}

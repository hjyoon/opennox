package server

import (
	"unsafe"

	noxflags "github.com/opennox/opennox/v1/common/flags"
	"github.com/opennox/opennox/v1/legacy/common/alloc"
)

const (
	objectExtendedDataFlagsMask4F40A0  = uint32(0x11408162)
	objectExtendedDataField5Mask4F40A0 = uint8(0x5e)
	objectExtendedDataModeMask4F40A0   = uint32(0x00600000)
	objectExtendedDataHostMask4F40A0   = uint32(0x00000001)
)

// objectExtendedDataDeps4F40A0 exposes every observable load and call in
// GAME.EXE 004F40A0. ID presence is deliberately separate from string
// contents: the original rejects every non-null ID pointer, including an
// empty C string.
type objectExtendedDataDeps4F40A0[O comparable, T any, S comparable] struct {
	loadIDPointerPresent func(O) bool
	loadInventoryHead    func(O) O
	loadField129         func(O) O
	loadTeamID           func(O) uint8
	loadTypeInd          func(O) uint16
	lookupType           func(uint16) T
	loadTypeFlags        func(T) uint32
	loadObjectFlags      func(O) uint32
	loadTypeField9       func(T) uint32
	loadObjectField5     func(O) uint32
	gameFlags            func(uint32) int32
	loadField189         func(O) S
	stringLength         func(S) uintptr
	loadScriptPickupFunc func(O) int32
}

// objectExtendedDataAdmission4F40A0 preserves the original admission order,
// live reloads, and low-byte 0/-1 return contract. Do not add guards after the
// entry null check: a missing type record and invalid subordinate pointers can
// fault at the same dependency boundary as GAME.EXE.
func objectExtendedDataAdmission4F40A0[O comparable, T any, S comparable](
	object O,
	deps objectExtendedDataDeps4F40A0[O, T, S],
) int8 {
	var zeroObject O
	if object == zeroObject {
		return 0
	}
	if deps.loadIDPointerPresent(object) {
		return -1
	}
	if deps.loadInventoryHead(object) != zeroObject {
		return -1
	}
	if deps.loadField129(object) != zeroObject {
		return -1
	}
	if deps.loadTeamID(object) != 0 {
		return -1
	}

	typeInd := deps.loadTypeInd(object)
	typ := deps.lookupType(typeInd)
	typeFlags := deps.loadTypeFlags(typ)
	objectFlags := deps.loadObjectFlags(object)
	if (typeFlags^objectFlags)&objectExtendedDataFlagsMask4F40A0 != 0 {
		return -1
	}
	typeField9 := deps.loadTypeField9(typ)
	objectField5 := deps.loadObjectField5(object)
	if (uint8(typeField9)^uint8(objectField5))&objectExtendedDataField5Mask4F40A0 != 0 {
		return -1
	}

	if deps.gameFlags(objectExtendedDataModeMask4F40A0) != 0 {
		text := deps.loadField189(object)
		var zeroText S
		if text != zeroText && deps.stringLength(text) != 0 {
			return -1
		}
		return 0
	}
	if deps.gameFlags(objectExtendedDataHostMask4F40A0) == 0 {
		return 0
	}
	if deps.loadScriptPickupFunc(object) != -1 {
		return -1
	}
	return 0
}

func (s *Server) Sub_4F40A0(object *Object) int8 {
	return objectExtendedDataAdmission4F40A0(object,
		objectExtendedDataDeps4F40A0[*Object, *ObjectType, unsafe.Pointer]{
			loadIDPointerPresent: func(object *Object) bool {
				return object.IDPtr != nil
			},
			loadInventoryHead: func(object *Object) *Object {
				return object.InvFirstItem
			},
			loadField129: func(object *Object) *Object {
				return object.Field129
			},
			loadTeamID: func(object *Object) uint8 {
				return uint8(object.TeamVal.ID)
			},
			loadTypeInd: func(object *Object) uint16 {
				return object.TypeInd
			},
			lookupType: func(typeInd uint16) *ObjectType {
				return s.Types.ByInd(int(typeInd))
			},
			loadTypeFlags: func(typ *ObjectType) uint32 {
				return uint32(typ.Flags())
			},
			loadObjectFlags: func(object *Object) uint32 {
				return uint32(object.ObjFlags)
			},
			loadTypeField9: func(typ *ObjectType) uint32 {
				return typ.Field9
			},
			loadObjectField5: func(object *Object) uint32 {
				return object.Field5
			},
			gameFlags: func(mask uint32) int32 {
				if noxflags.HasGame(noxflags.GameFlag(mask)) {
					return 1
				}
				return 0
			},
			loadField189: func(object *Object) unsafe.Pointer {
				return object.Field189
			},
			stringLength: func(text unsafe.Pointer) uintptr {
				return uintptr(len(alloc.GoString((*byte)(text))))
			},
			loadScriptPickupFunc: func(object *Object) int32 {
				return object.ScriptPickup.Func
			},
		})
}

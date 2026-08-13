package server

import "github.com/opennox/opennox/v1/common/ntype"

type objectPlayerMasksClearNativeDeps4E80C0 struct {
	firstObject func() *Object
	nextObject  func(*Object) *Object
}

func objectPlayerMasksClearNative4E80C0(
	playerInd uint32,
	deps objectPlayerMasksClearNativeDeps4E80C0,
) *Object {
	return objectPlayerMasksClear4E80C0(playerInd, objectPlayerMasksClearHooks4E80C0[*Object]{
		firstObject: deps.firstObject,
		loadField36: func(obj *Object) uint32 {
			return obj.Field36
		},
		loadField35: func(obj *Object) uint32 {
			return obj.Field35
		},
		storeField36: func(obj *Object, value uint32) {
			obj.Field36 = value
		},
		storeField35: func(obj *Object, value uint32) {
			obj.Field35 = value
		},
		nextObject: deps.nextObject,
	})
}

// ClearObjectPlayerMasks4E80C0 clears one player bit from both relationship
// masks of every active object. Conversion to uint32 retains the original
// IA-32 stack argument before its low-five-bit shift-count masking.
func (s *Server) ClearObjectPlayerMasks4E80C0(playerInd ntype.PlayerInd) {
	objectPlayerMasksClearNative4E80C0(uint32(playerInd), objectPlayerMasksClearNativeDeps4E80C0{
		firstObject: s.Objs.First,
		nextObject: func(obj *Object) *Object {
			return obj.Next()
		},
	})
}

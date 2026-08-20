package server

import (
	"unsafe"

	"github.com/opennox/libs/object"

	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// PlayerRespawnItemRuntime4EF750 supplies the two services whose current
// implementations remain outside server. Their Object and modifier pointers
// retain native width; placement scalars retain the original 32-bit C width.
type PlayerRespawnItemRuntime4EF750 struct {
	ApplyModifierAttrs func(*Object, *ModifierInitData)
	PlaceInventory     func(*Object, *Object, int32, int32) bool
}

type playerRespawnItemUpdatePrefix4EF750 struct {
	Field0 uint32
	Mark4  uint32
}

type playerRespawnItemNativeDeps4EF750 struct {
	newObject      func(string) *Object
	callInit       func(unsafe.Pointer, *Object, uint32)
	applyAttrs     func(*Object, *ModifierInitData)
	placeInventory func(*Object, *Object, int32, int32) bool
}

func playerRespawnItemNative4EF750(
	player *Object,
	typeID string,
	attrs *ModifierInitData,
	a4 int32,
	a5 int32,
	deps playerRespawnItemNativeDeps4EF750,
) *Object {
	return playerRespawnItem4EF750(playerRespawnItemHooks4EF750[
		*Object,
		unsafe.Pointer,
		*ModifierInitData,
		*playerRespawnItemUpdatePrefix4EF750,
	]{
		loadTypeIDArg: func() string {
			return typeID
		},
		newObject: deps.newObject,
		loadInit: func(item *Object) unsafe.Pointer {
			return item.Init
		},
		callInit: deps.callInit,
		loadAttrsArg: func() *ModifierInitData {
			return attrs
		},
		applyAttrs: deps.applyAttrs,
		loadPlaceA5Arg: func() int32 {
			return a5
		},
		loadPlaceA4Arg: func() int32 {
			return a4
		},
		loadPlayerArg: func() *Object {
			return player
		},
		placeInventory: func(player, item *Object, a4, a5 int32) {
			_ = deps.placeInventory(player, item, a4, a5)
		},
		loadFlags: func(item *Object) uint32 {
			return uint32(item.ObjFlags)
		},
		loadClass: func(item *Object) uint32 {
			return uint32(item.ObjClass)
		},
		storeFlags: func(item *Object, flags uint32) {
			item.ObjFlags = object.Flags(flags)
		},
		loadUpdateData: func(item *Object) *playerRespawnItemUpdatePrefix4EF750 {
			return (*playerRespawnItemUpdatePrefix4EF750)(item.UpdateData)
		},
		loadUpdateMark: func(update *playerRespawnItemUpdatePrefix4EF750) uint32 {
			return update.Mark4
		},
		storeUpdateMark: func(update *playerRespawnItemUpdatePrefix4EF750, mark uint32) {
			update.Mark4 = mark
		},
	})
}

func playerRespawnItemServerDeps4EF750(
	s *Server,
	runtime PlayerRespawnItemRuntime4EF750,
) playerRespawnItemNativeDeps4EF750 {
	return playerRespawnItemNativeDeps4EF750{
		newObject: s.NewObjectByTypeID,
		callInit: func(init unsafe.Pointer, item *Object, value uint32) {
			ccall.CallVoidUPtr2(init, uintptr(item.CObj()), uintptr(value))
		},
		applyAttrs:     runtime.ApplyModifierAttrs,
		placeInventory: runtime.PlaceInventory,
	}
}

// PlayerRespawnItem4EF750 binds GAME.EXE 004EF750 to native-width Object,
// initializer, modifier, and UpdateData pointers. Object flags, class bits,
// the UpdateData mark at byte offset four, and placement scalars retain their
// original fixed 32-bit widths.
func (s *Server) PlayerRespawnItem4EF750(
	player *Object,
	typeID string,
	attrs *ModifierInitData,
	a4 int32,
	a5 int32,
	runtime PlayerRespawnItemRuntime4EF750,
) *Object {
	return playerRespawnItemNative4EF750(
		player,
		typeID,
		attrs,
		a4,
		a5,
		playerRespawnItemServerDeps4EF750(s, runtime),
	)
}

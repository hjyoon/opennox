package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// WinkGameBallReleaseRuntime4F7DF0 supplies services that remain owned by the
// root runtime. Object pointers stay native-width; only the original status
// arguments and the fixed-width object fields are narrowed.
type WinkGameBallReleaseRuntime4F7DF0 struct {
	ApplyForce func(*Object, types.Pointf, float64)
	BallStatus func(uint8, uint16) int32
}

type winkGameBallReleaseNativeDeps4F7DF0 struct {
	loadTypeCache  func() uint32
	lookupType     func(string) uint32
	storeTypeCache func(uint32)
	applyForce     func(*Object, *Object, float32)
	clearOwner     func(*Object)
	audio          func(uint32, *Object, int32, uint32)
	ballStatus     func(uint8, uint16) int32
}

func winkGameBallReleaseNative4F7DF0(
	player *Object,
	deps winkGameBallReleaseNativeDeps4F7DF0,
) int32 {
	return winkGameBallRelease4F7DF0(player, winkGameBallReleaseHooks4F7DF0[*Object]{
		loadTypeCache:  deps.loadTypeCache,
		lookupType:     deps.lookupType,
		storeTypeCache: deps.storeTypeCache,
		loadFirstOwned: func(owner *Object) *Object {
			return owner.Field129
		},
		loadTypeInd: func(obj *Object) uint16 {
			return obj.TypeInd
		},
		loadNextOwned: func(obj *Object) *Object {
			return obj.Field128
		},
		loadFlags: func(obj *Object) uint32 {
			return uint32(obj.ObjFlags)
		},
		storeFlags: func(obj *Object, flags uint32) {
			obj.ObjFlags = object.Flags(flags)
		},
		applyForce: deps.applyForce,
		storeObj130: func(obj, value *Object) {
			obj.Obj130 = value
		},
		clearOwner: deps.clearOwner,
		audio:      deps.audio,
		ballStatus: deps.ballStatus,
	})
}

func winkGameBallReleaseServerDeps4F7DF0(
	s *Server,
	runtime WinkGameBallReleaseRuntime4F7DF0,
) winkGameBallReleaseNativeDeps4F7DF0 {
	return winkGameBallReleaseNativeDeps4F7DF0{
		loadTypeCache: func() uint32 {
			return s.Types.fast.winkGameBall4F7DF0
		},
		lookupType: func(name string) uint32 {
			return uint32(s.Types.IndByID(name))
		},
		storeTypeCache: func(value uint32) {
			s.Types.fast.winkGameBall4F7DF0 = value
		},
		applyForce: func(player, ball *Object, force float32) {
			runtime.ApplyForce(ball, player.PosVec, float64(force))
		},
		clearOwner: s.ObjClearOwner,
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
		ballStatus: runtime.BallStatus,
	}
}

// WinkGameBallRelease4F7DF0 binds GAME.EXE 004F7DF0 to native-width Object
// fields and its private fixed-width type cache. It intentionally retains the
// original nil-player fault behavior after cache initialization.
func (s *Server) WinkGameBallRelease4F7DF0(
	player *Object,
	runtime WinkGameBallReleaseRuntime4F7DF0,
) int32 {
	return winkGameBallReleaseNative4F7DF0(
		player,
		winkGameBallReleaseServerDeps4F7DF0(s, runtime),
	)
}

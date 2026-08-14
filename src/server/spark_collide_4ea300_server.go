package server

import (
	"unsafe"

	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// SparkUpdateData is the fixed-width 16-byte record allocated for SparkUpdate.
// Collide uses only Kind; the existing update routine also consumes Field4.
type SparkUpdateData struct {
	Field0 uint32
	Field4 uint32
	Field8 uint32
	Kind   uint32
}

// SparkCollideRuntime4EA300 supplies the legacy-owned map and deletion effects
// shared with WallReflectCollide. Native object, update-data, audio, and
// priority-message state remain inside the server adapter.
type SparkCollideRuntime4EA300 struct {
	WallReflect WallReflectCollideRuntime4E9D80
}

type sparkCollideNativeDeps4EA300 struct {
	wallReflect     func(*Object, *Object, *types.Pointf)
	audio           func(uint32, *Object)
	delayedDelete   func(*Object)
	priorityMessage func(*Object, string)
}

func sparkCollideNative4EA300(
	source, target *Object,
	collision *types.Pointf,
	deps sparkCollideNativeDeps4EA300,
) {
	sparkCollide4EA300(source, target, collision, sparkCollideHooks4EA300[
		*Object,
		*types.Pointf,
		*SparkUpdateData,
	]{
		loadUpdateData: func(obj *Object) *SparkUpdateData {
			return (*SparkUpdateData)(obj.UpdateData)
		},
		loadKind: func(data *SparkUpdateData) uint32 {
			return data.Kind
		},
		wallReflect:   deps.wallReflect,
		audio:         deps.audio,
		delayedDelete: deps.delayedDelete,
		loadSlowCount: func(obj *Object) uint8 {
			return obj.Field541
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		storeSlowCount: func(obj *Object, count uint8) {
			obj.Field541 = count
		},
		storeSlowTimer: func(obj *Object, timer uint16) {
			obj.Field542 = timer
		},
		priorityMessage: deps.priorityMessage,
	})
}

func sparkCollideServerDeps4EA300(
	s *Server,
	runtime SparkCollideRuntime4EA300,
) sparkCollideNativeDeps4EA300 {
	wallDeps := wallReflectNativeDeps4E9D80(s, runtime.WallReflect)
	return sparkCollideNativeDeps4EA300{
		wallReflect: func(source, target *Object, collision *types.Pointf) {
			wallReflectCollideNative4E9D80(source, target, collision, wallDeps)
		},
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		delayedDelete: runtime.WallReflect.DelayedDelete,
		priorityMessage: func(obj *Object, message string) {
			s.NetPriMsgToPlayer(obj, strman.ID(message), 0)
		},
	}
}

// SparkCollide4EA300 binds the original callback to native Object pointers,
// the fixed-width Spark update record, and the restored WallReflect callback.
func (s *Server) SparkCollide4EA300(
	source, target *Object,
	collision *types.Pointf,
	runtime SparkCollideRuntime4EA300,
) {
	sparkCollideNative4EA300(source, target, collision, sparkCollideServerDeps4EA300(s, runtime))
}

var (
	_ = [1]struct{}{}[16-unsafe.Sizeof(SparkUpdateData{})]
	_ = [1]struct{}{}[12-unsafe.Offsetof(SparkUpdateData{}.Kind)]
)

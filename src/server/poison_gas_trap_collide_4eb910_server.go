package server

import (
	"unsafe"

	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// ToxicCloudUpdateData is the fixed-width four-byte record registered for
// ToxicCloudUpdate and SmallToxicCloudUpdate. PoisonGasTrapCollide writes the
// lifetime, in frames, to its first and only dword.
type ToxicCloudUpdateData struct {
	Duration int32
}

// PoisonGasTrapCollideRuntime4EB910 supplies object lifecycle effects that
// remain owned by the legacy-facing runtime. The trap gate, balance, frame
// rate, native Object fields, update record, conversion, and audio stay in
// package server.
type PoisonGasTrapCollideRuntime4EB910 struct {
	CreateAt      func(*Object, *Object, types.Pointf)
	DelayedDelete func(*Object)
}

type poisonGasTrapCollideNativeDeps4EB910 struct {
	allowed       func(*Object, *Object) int32
	newObject     func(string) *Object
	createAt      func(*Object, *Object, types.Pointf, uint32)
	loadLifetime  func(string) float32
	loadFPS       func() uint32
	multiply      func(float32, uint32) float32
	floatToInt    func(float32) int32
	audio         func(uint32, *Object, int32, uint32)
	delayedDelete func(*Object)
}

func poisonGasTrapCollideNative4EB910(
	source, target *Object,
	collision *types.Pointf,
	deps poisonGasTrapCollideNativeDeps4EB910,
) {
	poisonGasTrapCollide4EB910(source, target, collision, poisonGasTrapCollideHooks4EB910[
		*Object,
		*ToxicCloudUpdateData,
	]{
		allowed:   deps.allowed,
		newObject: deps.newObject,
		loadPosY: func(obj *Object) float32 {
			return obj.PosVec.Y
		},
		loadPosX: func(obj *Object) float32 {
			return obj.PosVec.X
		},
		loadOwner: func(obj *Object) *Object {
			return obj.ObjOwner
		},
		createAt: func(obj, owner *Object, x, y float32, reserved uint32) {
			deps.createAt(obj, owner, types.Pointf{X: x, Y: y}, reserved)
		},
		loadUpdateData: func(obj *Object) *ToxicCloudUpdateData {
			return (*ToxicCloudUpdateData)(obj.UpdateData)
		},
		loadLifetime: deps.loadLifetime,
		loadFPS:      deps.loadFPS,
		multiply:     deps.multiply,
		floatToInt:   deps.floatToInt,
		storeDuration: func(data *ToxicCloudUpdateData, duration int32) {
			data.Duration = duration
		},
		audio:         deps.audio,
		delayedDelete: deps.delayedDelete,
	})
}

func poisonGasTrapCollideServerDeps4EB910(
	s *Server,
	runtime PoisonGasTrapCollideRuntime4EB910,
) poisonGasTrapCollideNativeDeps4EB910 {
	return poisonGasTrapCollideNativeDeps4EB910{
		allowed:   s.GlyphCollideAllowed4E9A30,
		newObject: s.NewObjectByTypeID,
		createAt: func(obj, owner *Object, pos types.Pointf, _ uint32) {
			runtime.CreateAt(obj, owner, pos)
		},
		loadLifetime: func(key string) float32 {
			return float32(s.Balance.Float(key))
		},
		loadFPS:    s.TickRate,
		multiply:   poisonGasTrapMultiply4EB910,
		floatToInt: poisonGasTrapRound4EB910,
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
		delayedDelete: runtime.DelayedDelete,
	}
}

// PoisonGasTrapCollide4EB910 binds the zero-byte PoisonGasTrapCollide
// registration to native-width Object pointers and the fixed-width ToxicCloud
// update record while preserving GAME.EXE field-read and effect order.
func (s *Server) PoisonGasTrapCollide4EB910(
	source, target *Object,
	collision *types.Pointf,
	runtime PoisonGasTrapCollideRuntime4EB910,
) {
	poisonGasTrapCollideNative4EB910(source, target, collision, poisonGasTrapCollideServerDeps4EB910(s, runtime))
}

var (
	_ = [1]struct{}{}[4-unsafe.Sizeof(ToxicCloudUpdateData{})]
	_ = [1]struct{}{}[0-unsafe.Offsetof(ToxicCloudUpdateData{}.Duration)]
)

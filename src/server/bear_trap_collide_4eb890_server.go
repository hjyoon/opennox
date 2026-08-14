package server

import (
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
)

// BearTrapCollideRuntime4EB890 supplies the object-lifecycle and enchantment
// effects that remain owned by the legacy-facing runtime. The shared trap
// gate, object lookup, native Object fields, and audio stay in package server.
type BearTrapCollideRuntime4EB890 struct {
	CreateAt      func(*Object, *Object, types.Pointf)
	DelayedDelete func(*Object)
	ApplyEnchant  func(*Object, EnchantID, uint32, uint32)
}

type bearTrapCollideNativeDeps4EB890 struct {
	allowed       func(*Object, *Object) int32
	newObject     func(string) *Object
	createAt      func(*Object, *Object, types.Pointf, uint32)
	delayedDelete func(*Object)
	applyEnchant  func(*Object, EnchantID, uint32, uint32)
	audio         func(uint32, *Object, int32, uint32)
}

func bearTrapCollideNative4EB890(
	source, target *Object,
	collision *types.Pointf,
	deps bearTrapCollideNativeDeps4EB890,
) {
	bearTrapCollide4EB890(source, target, collision, bearTrapCollideHooks4EB890[*Object]{
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
		delayedDelete: deps.delayedDelete,
		applyEnchant: func(obj *Object, enchant, duration, power uint32) {
			deps.applyEnchant(obj, EnchantID(enchant), duration, power)
		},
		audio: deps.audio,
	})
}

func bearTrapCollideServerDeps4EB890(
	s *Server,
	runtime BearTrapCollideRuntime4EB890,
) bearTrapCollideNativeDeps4EB890 {
	return bearTrapCollideNativeDeps4EB890{
		allowed:   s.GlyphCollideAllowed4E9A30,
		newObject: s.NewObjectByTypeID,
		createAt: func(obj, owner *Object, pos types.Pointf, _ uint32) {
			runtime.CreateAt(obj, owner, pos)
		},
		delayedDelete: runtime.DelayedDelete,
		applyEnchant:  runtime.ApplyEnchant,
		audio: func(id uint32, obj *Object, kind int32, code uint32) {
			s.Audio.EventObj(sound.ID(id), obj, int(kind), code)
		},
	}
}

// BearTrapCollide4EB890 binds the zero-byte BearTrapCollide registration to
// native-width Object pointers while preserving GAME.EXE field-read and
// effect order.
func (s *Server) BearTrapCollide4EB890(
	source, target *Object,
	collision *types.Pointf,
	runtime BearTrapCollideRuntime4EB890,
) {
	bearTrapCollideNative4EB890(source, target, collision, bearTrapCollideServerDeps4EB890(s, runtime))
}

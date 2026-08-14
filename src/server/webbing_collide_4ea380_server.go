package server

import (
	"github.com/opennox/libs/object"
	"github.com/opennox/libs/strman"
	"github.com/opennox/libs/types"

	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/legacy/common/ccall"
)

// WebbingCollideRuntime4EA380 supplies the object-lifecycle and enchantment
// effects that remain implemented by the legacy bridge. Native object fields,
// audio, frame rate, damage dispatch, and priority messages stay in the server.
type WebbingCollideRuntime4EA380 struct {
	DelayedDelete func(*Object)
	ApplyEnchant  func(*Object, EnchantID, uint32, uint32)
}

type webbingCollideNativeDeps4EA380 struct {
	audio           func(uint32, *Object)
	delayedDelete   func(*Object)
	findParent      func(*Object) *Object
	targetDamage    func(*Object, *Object, *Object, int32, object.DamageType) int32
	loadFPS         func() uint32
	applyEnchant    func(*Object, EnchantID, uint32, uint32)
	priorityMessage func(*Object, string)
}

func webbingCollideNative4EA380(
	source, target *Object,
	collision *types.Pointf,
	deps webbingCollideNativeDeps4EA380,
) {
	webbingCollide4EA380(source, target, collision, webbingCollideHooks4EA380[*Object]{
		audio:         deps.audio,
		delayedDelete: deps.delayedDelete,
		findParent:    deps.findParent,
		targetDamage: func(target, parent, source *Object, damage int32, damageType uint32) int32 {
			return deps.targetDamage(target, parent, source, damage, object.DamageType(damageType))
		},
		loadClassLow: func(obj *Object) uint8 {
			return uint8(obj.ObjClass)
		},
		loadFPS: deps.loadFPS,
		applyEnchant: func(obj *Object, enchant, duration, power uint32) {
			deps.applyEnchant(obj, EnchantID(enchant), duration, power)
		},
		priorityMessage: deps.priorityMessage,
	})
}

func webbingCollideServerDeps4EA380(
	s *Server,
	runtime WebbingCollideRuntime4EA380,
) webbingCollideNativeDeps4EA380 {
	return webbingCollideNativeDeps4EA380{
		audio: func(id uint32, obj *Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		delayedDelete: runtime.DelayedDelete,
		findParent:    (*Object).FindOwnerChainPlayer,
		targetDamage: func(target, parent, source *Object, damage int32, damageType object.DamageType) int32 {
			return int32(ccall.CallIntUPtr5(
				target.Damage,
				uintptr(target.CObj()),
				uintptr(toObjectC(parent)),
				uintptr(toObjectC(source)),
				uintptr(uint32(damage)),
				uintptr(uint32(damageType)),
			))
		},
		loadFPS:      s.TickRate,
		applyEnchant: runtime.ApplyEnchant,
		priorityMessage: func(obj *Object, message string) {
			s.NetPriMsgToPlayer(obj, strman.ID(message), 0)
		},
	}
}

// WebbingCollide4EA380 binds the SpiderSpitCollide registration to the native
// Object layout while preserving the original callback order and live loads.
func (s *Server) WebbingCollide4EA380(
	source, target *Object,
	collision *types.Pointf,
	runtime WebbingCollideRuntime4EA380,
) {
	webbingCollideNative4EA380(source, target, collision, webbingCollideServerDeps4EA380(s, runtime))
}

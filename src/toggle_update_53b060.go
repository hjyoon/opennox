package opennox

import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) updateToggleNative53B060(unit *server.Object) {
	s.ToggleUpdate53B060(unit, server.ToggleUpdateRuntime53B060{
		CollideTarget: func(obj *server.Object) *server.Object {
			return obj.TriggerCollideTarget()
		},
		AudioEvent: func(id uint32, obj *server.Object) {
			s.Audio.EventObj(sound.ID(id), obj, 0, 0)
		},
		ScriptCallback: func(block *server.ScriptCallback, caller, trigger *server.Object, event server.ScriptEventType) {
			s.noxScript.ScriptCallback(block, caller, trigger, event)
		},
	})
}

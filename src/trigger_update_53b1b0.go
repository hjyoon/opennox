package opennox

import (
	"github.com/opennox/opennox/v1/common/sound"
	"github.com/opennox/opennox/v1/server"
)

func (s *Server) updateTriggerNative53B1B0(unit *server.Object) {
	s.TriggerUpdate53B1B0(unit, server.TriggerUpdateRuntime53B1B0{
		ImmediateType: func(obj *server.Object) bool {
			trigger := s.Types.ByID("Trigger")
			plate := s.Types.ByID("PressurePlate")
			return trigger != nil && obj.TypeInd == uint16(trigger.Ind()) ||
				plate != nil && obj.TypeInd == uint16(plate.Ind())
		},
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

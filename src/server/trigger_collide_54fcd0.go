package server

import "github.com/opennox/libs/object"

type TriggerCollideRuntime54FCD0 struct {
	Mass          func(*Object) float64
	ScriptAllowed func(*ScriptCallback, *Object, *Object, ScriptEventType) bool
}

// TriggerCollide54FCD0 ports GAME.EXE 0054FCD0 while keeping the current
// collide target in ObjectExt instead of TriggerUpdateData's PE32 Field4.
func (s *Server) TriggerCollide54FCD0(trigger, candidate *Object, runtime TriggerCollideRuntime54FCD0) {
	if trigger == nil || trigger.UpdateData == nil {
		return
	}
	update := trigger.UpdateDataTrigger()
	if !trigger.ObjFlags.Has(object.FlagEnabled) {
		trigger.SetTriggerCollideTarget(nil)
		update.Flags &^= 1
		return
	}
	if update.State == 5 || candidate == nil || runtime.Mass(candidate) <= 0 {
		return
	}
	class := uint32(candidate.ObjClass)
	if update.ClassInclude != 0 && update.ClassInclude&class == 0 {
		return
	}
	if update.ClassExclude != 0 && update.ClassExclude&class != 0 {
		return
	}
	team := uint8(candidate.TeamVal.ID)
	if update.TeamInclude != 0 && update.TeamInclude != team {
		return
	}
	if update.TeamExclude != 0 && update.TeamExclude == team {
		return
	}
	if update.ScriptCollide.Func != -1 && !runtime.ScriptAllowed(
		&update.ScriptCollide, candidate, trigger, NoxEventCollide,
	) {
		return
	}
	trigger.SetTriggerCollideTarget(candidate)
	update.Flags |= 1
}

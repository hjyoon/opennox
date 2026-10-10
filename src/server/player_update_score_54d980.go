package server

import (
	"unsafe"

	"github.com/opennox/libs/object"
)

// These callbacks are the external service boundaries of GAME.EXE 0054D980.
// The root caches update/team pointers before callbacks, not their live fields.
type PlayerUpdateScoreRuntime54D980 struct {
	HasTeam            func(*ObjectTeam) bool
	TeamByID           func(uint8) *Team
	AddScore           func(*Object, uint32)
	SubtractScore      func(*Object, uint32)
	ReportLesson       func(*Object)
	IncrementElimDeath func(*Object)
	TeamChangeLessons  func(*Team, int32)
	ObserverMode       func() uint32
	ObserverUpdate     func(*Player, *Player)
}

// PlayerScoreAdd4D8E90 preserves the original DWORD addition and return identity.
// Even a non-player loads its update pointer before the class test; only a
// player dereferences that pointer. There is no additional nil/update gate.
func PlayerScoreAdd4D8E90(unit *Object, value uint32) unsafe.Pointer {
	class := unit.ObjClass
	update := (*PlayerUpdateData)(unit.UpdateData)
	if class.Has(object.ClassPlayer) {
		player := update.Player
		player.Lessons = int32(uint32(player.Lessons) + value)
		return unsafe.Pointer(player)
	}
	return unsafe.Pointer(unit)
}

// PlayerScoreSubtract4D8EC0 is the corresponding original DWORD subtraction.
func PlayerScoreSubtract4D8EC0(unit *Object, value uint32) unsafe.Pointer {
	class := unit.ObjClass
	update := (*PlayerUpdateData)(unit.UpdateData)
	if class.Has(object.ClassPlayer) {
		player := update.Player
		player.Lessons = int32(uint32(player.Lessons) - value)
		return unsafe.Pointer(player)
	}
	return unsafe.Pointer(unit)
}

// PlayerUpdateScore54D980 follows the executed PE32 bytes, not the additional
// nil-team guard in the decompilation. It is a scoring dependency, not admission
// of the still-unported competitive PlayerDie branches. Tracking is a DWORD
// nonzero test; an assist is awarded independently of its player class.
func PlayerUpdateScore54D980(victim, killer, assist *Object, tracking uint32, rt PlayerUpdateScoreRuntime54D980) {
	victimUpdate := (*PlayerUpdateData)(victim.UpdateData)
	var victimTeam, killerTeam, assistTeam *Team
	var killerUpdate, assistUpdate *PlayerUpdateData
	if rt.HasTeam(&victim.TeamVal) {
		victimTeam = rt.TeamByID(uint8(victim.TeamVal.ID))
	}
	if killer != nil {
		killerUpdate = (*PlayerUpdateData)(killer.UpdateData)
		if rt.HasTeam(&killer.TeamVal) {
			killerTeam = rt.TeamByID(uint8(killer.TeamVal.ID))
		}
	}
	if tracking != 0 && assist != nil {
		assistUpdate = (*PlayerUpdateData)(assist.UpdateData)
		if rt.HasTeam(&assist.TeamVal) {
			assistTeam = rt.TeamByID(uint8(assist.TeamVal.ID))
		}
	}

	if killer == victim || killer == nil && assist == nil {
		rt.SubtractScore(victim, 1)
		rt.ReportLesson(victim)
		if victimTeam != nil {
			rt.TeamChangeLessons(victimTeam, int32(uint32(victimTeam.Lessons)-1))
		}
		if rt.ObserverMode() != 0 && killerUpdate != nil {
			player := killerUpdate.Player
			rt.ObserverUpdate(player, player)
		}
	} else {
		if killer != nil && killer.ObjClass.Has(object.ClassPlayer) {
			if killerTeam != nil && killerTeam == victimTeam {
				rt.SubtractScore(killer, 1)
				rt.ReportLesson(killer)
				rt.TeamChangeLessons(killerTeam, int32(uint32(killerTeam.Lessons)-1))
				if rt.ObserverMode() != 0 && killerUpdate != nil {
					player := killerUpdate.Player
					rt.ObserverUpdate(player, player)
				}
			} else if killerTeam == nil && victimTeam == nil {
				rt.AddScore(killer, 1)
				rt.ReportLesson(killer)
				if rt.ObserverMode() != 0 && killerUpdate != nil && victimUpdate != nil {
					// 0054DA9A loads victim's Player before killer's Player.
					second := victimUpdate.Player
					first := killerUpdate.Player
					rt.ObserverUpdate(first, second)
				}
			} else {
				rt.AddScore(killer, 1)
				rt.ReportLesson(killer)
				// 0054DABB reads this even when only the victim has a team.
				rt.TeamChangeLessons(killerTeam, int32(uint32(killerTeam.Lessons)+1))
				if rt.ObserverMode() != 0 && killerUpdate != nil && victimUpdate != nil {
					first := killerUpdate.Player
					second := victimUpdate.Player
					rt.ObserverUpdate(first, second)
				}
			}
		}
		rt.IncrementElimDeath(victim)
		rt.ReportLesson(victim)
	}

	if assist == nil {
		return
	}
	if killerTeam != nil {
		if killerTeam == assistTeam {
			return
		}
	} else if assistTeam == nil {
		// 0054DB9A skips the victim-team comparison for two nil teams.
		rt.AddScore(assist, 1)
		rt.ReportLesson(assist)
		if rt.ObserverMode() != 0 && assistUpdate != nil && victimUpdate != nil {
			second := victimUpdate.Player
			first := assistUpdate.Player
			rt.ObserverUpdate(first, second)
		}
		return
	}
	if victimTeam != nil && victimTeam == assistTeam {
		return
	}
	rt.AddScore(assist, 1)
	rt.ReportLesson(assist)
	if assistTeam != nil {
		rt.TeamChangeLessons(assistTeam, int32(uint32(assistTeam.Lessons)+1))
	}
	if rt.ObserverMode() != 0 && assistUpdate != nil && victimUpdate != nil {
		second := victimUpdate.Player
		first := assistUpdate.Player
		rt.ObserverUpdate(first, second)
	}
}
